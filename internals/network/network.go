package network

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

func AssignIP(link netlink.Link, ip string) error {
	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(link, addr)
}

func createBridge() (*netlink.Bridge, error) {
	bridge := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: "cage0",
		},
	}

	if err := netlink.LinkAdd(bridge); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(bridge); err != nil {
		return nil, err
	}
	if err := AssignIP(bridge, "10.0.0.1/24"); err != nil {
		return nil, err
	}

	return bridge, nil
}

func GetorCreateBridge() (*netlink.Bridge, error) {
	link, err := netlink.LinkByName("cage0")
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			return createBridge()
		}
		return nil, err
	}

	bridge, ok := link.(*netlink.Bridge)
	if !ok {
		return nil, fmt.Errorf("cage0 exists but is not a bridge")
	}
	return bridge, nil
}

func ConnectToBridge(hostname string, bridge *netlink.Bridge) error {
	link, err := netlink.LinkByName(hostname)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetMaster(link, bridge); err != nil {
		return err
	}
	return nil
}

func CreateVethPair(hostname, peername string) error {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: hostname,
		},
		PeerName: peername,
	}

	return netlink.LinkAdd(veth)
}

func ConnectContainer(hostname, peername string, containerPID int) error {
	host, err := netlink.LinkByName(hostname)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetUp(host); err != nil {
		return err
	}
	peer, err := netlink.LinkByName(peername)
	if err != nil {
		return err
	}
	return netlink.LinkSetNsPid(peer, containerPID)
}

func getIPsInNamespace(nsHandler netns.NsHandle) ([]string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// gets the current network namespace and switch to the target namespace
	origns, err := netns.Get()
	if err != nil {
		return nil, err
	}
	defer origns.Close()
	// origin namespace is from the host
	// ns handler is from the container
	if err := netns.Set(nsHandler); err != nil {
		return nil, err
	}
	// set the namespace back to the host after this function returns
	defer netns.Set(origns)
	// get the list of network interfaces and their IP addresses in the container namespace
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, link := range links {
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ips = append(ips, addr.IP.String())
		}
	}

	return ips, nil
}

func FindFreeIP() (string, error) {
	// get the list of all veth-host* interfaces and their associated container IPs to find a free IP address in the
	links, err := netlink.LinkList()
	if err != nil {
		return "", err
	}

	usedIPs := make(map[string]bool)
	usedIPs["10.0.0.1"] = true

	for _, link := range links {
		name := link.Attrs().Name
		if len(name) > 6 && name[:6] == "veth-h" {
			pidStr := name[6:]
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				continue
			}

			nsHandler, err := netns.GetFromPid(pid)
			if err != nil {
				continue
			}

			ips, err := getIPsInNamespace(nsHandler)
			nsHandler.Close()
			if err != nil {
				continue
			}

			for _, ip := range ips {
				usedIPs[ip] = true
			}
		}
	}

	for i := 2; i < 255; i++ {
		ipStr := fmt.Sprintf("10.0.0.%d", i)
		if !usedIPs[ipStr] {
			return ipStr + "/24", nil
		}
	}

	return "", fmt.Errorf("no free IP addresses in 10.0.0.0/24 subnet")
}

func SetUpVeth(peername string, containerIP string) error {
	links, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list network interfaces: %w", err)
	}
	// find the veth link that starts with "veth-g" which is the peer in the container namespace
	var link netlink.Link
	for _, l := range links {
		name := l.Attrs().Name
		if len(name) > 6 && name[:6] == "veth-g" {
			link = l
			break
		}
	}

	if link == nil {
		return fmt.Errorf("failed to find any guest veth link starting with 'veth-g'")
	}

	_ = netlink.LinkSetDown(link)
	// chanfe veth-g to eth0
	if err := netlink.LinkSetName(link, peername); err != nil {
		return fmt.Errorf("failed to rename link to %s: %w", peername, err)
	}

	link, err = netlink.LinkByName(peername)
	if err != nil {
		return fmt.Errorf("failed to find renamed link %s: %w", peername, err)
	}

	if err := AssignIP(link, containerIP); err != nil {
		return fmt.Errorf("failed to assign IP %s to %s: %w", containerIP, peername, err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to set link %s up: %w", peername, err)
	}

	lo, err := netlink.LinkByName("lo")
	if err == nil {
		_ = netlink.LinkSetUp(lo)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        net.ParseIP("10.0.0.1"),
	}

	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("failed to add default route: %w", err)
	}

	return nil
}

func SetUpContainerNetwork(containerPid int, bridge *netlink.Bridge, hostnet string) error {
	guestnet := "veth-g" + strconv.Itoa(containerPid)
	if err := CreateVethPair(hostnet, guestnet); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}
	//connect host to bridge
	if err := ConnectToBridge(hostnet, bridge); err != nil {
		return fmt.Errorf("failed to connect host to bridge: %w", err)
	}
	// move peer to container
	if err := ConnectContainer(hostnet, guestnet, containerPid); err != nil {
		return fmt.Errorf("failed to move peer to container: %w", err)
	}

	if err := NftableSetup(); err != nil {
		return fmt.Errorf("failed to setup nftables: %w", err)
	}
	return nil
}

func CleanBridge(hostnet string) error {
	link, err := netlink.LinkByName(hostnet)
	if err != nil {
		return err
	}
	err = netlink.LinkDel(link)
	if err != nil {
		return err
	}
	return nil
}

func iptablesSetup() {
	// Add rules to iptables to allow forwarding on cage0 in case Docker or UFW has a default DROP policy.
	// We ignore errors since iptables might not be installed or available.
	_ = exec.Command("iptables", "-I", "FORWARD", "-i", "cage0", "-j", "ACCEPT").Run()
	_ = exec.Command("iptables", "-I", "FORWARD", "-o", "cage0", "-j", "ACCEPT").Run()
}

func tableExists(conn *nftables.Conn, family nftables.TableFamily, name string) (bool, error) {
	tables, err := conn.ListTables()
	if err != nil {
		return false, err
	}

	for _, t := range tables {
		if t.Name == name && t.Family == family {
			return true, nil
		}
	}
	return false, nil
}

func ifnameBytes(name string) []byte {
	b := make([]byte, unix.IFNAMSIZ)
	copy(b, name)
	return b
}

func NftableSetup() error {
	conn, err := nftables.New()
	if err != nil {
		return errors.New("failed to create nftables connection: " + err.Error())
	}

	natExists, err := tableExists(conn, nftables.TableFamilyIPv4, "cage_nat")
	if err == nil && natExists {
		conn.DelTable(&nftables.Table{
			Family: nftables.TableFamilyIPv4,
			Name:   "cage_nat",
		})
	}
	filterExists, err := tableExists(conn, nftables.TableFamilyINet, "cage_filter")
	if err == nil && filterExists {
		conn.DelTable(&nftables.Table{
			Family: nftables.TableFamilyINet,
			Name:   "cage_filter",
		})
	}
	if (err == nil && natExists) || (err == nil && filterExists) {
		_ = conn.Flush()
	}

	natTable := conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyIPv4,
		Name:   "cage_nat",
	})
	postrouting := conn.AddChain(&nftables.Chain{
		Name:     "postrouting",
		Table:    natTable,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})
	_, subnet, err := net.ParseCIDR("10.0.0.0/24")
	if err != nil {
		return errors.New("failed to parse subnet: " + err.Error())
	}
	conn.AddRule(&nftables.Rule{
		Table: natTable,
		Chain: postrouting,
		Exprs: []expr.Any{
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       12, // IPv4 source address
				Len:          4,
			},

			// Match 10.0.0.0/24
			&expr.Bitwise{
				SourceRegister: 1,
				DestRegister:   1,
				Len:            4,
				Mask:           []byte{255, 255, 255, 0},
				Xor:            []byte{0, 0, 0, 0},
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     subnet.IP.To4(),
			},

			// Load output interface name into register 1
			&expr.Meta{
				Key:      expr.MetaKeyOIFNAME,
				Register: 1,
			},

			// Match outgoing interface != "cage0"
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpNeq,
				Data:     ifnameBytes("cage0"),
			},

			&expr.Masq{},
		},
	})
	filterTable := conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "cage_filter",
	})
	policy := nftables.ChainPolicyAccept
	forwarding := conn.AddChain(&nftables.Chain{
		Name:     "forwarding",
		Table:    filterTable,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Policy:   &policy,
		Priority: nftables.ChainPriorityFilter,
	})

	// forwarding: iifname "cage0" accept
	conn.AddRule(&nftables.Rule{
		Table: filterTable,
		Chain: forwarding,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyIIFNAME,
				Register: 1,
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     ifnameBytes("cage0"),
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	// forwarding: oifname "cage0" ct state established,related accept
	conn.AddRule(&nftables.Rule{
		Table: filterTable,
		Chain: forwarding,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyOIFNAME,
				Register: 1,
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     ifnameBytes("cage0"),
			},
			// ct state established,related
			&expr.Ct{
				Key:      expr.CtKeySTATE,
				Register: 1,
			},
			&expr.Bitwise{
				SourceRegister: 1,
				DestRegister:   1,
				Len:            4,
				Mask: []byte{
					byte(expr.CtStateBitESTABLISHED |
						expr.CtStateBitRELATED),
					0, 0, 0,
				},
				Xor: []byte{0, 0, 0, 0},
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpNeq,
				Data:     []byte{0, 0, 0, 0},
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	if err := conn.Flush(); err != nil {
		return errors.New("failed to flush nftables rules: " + err.Error())
	}
	if err := os.WriteFile(
		"/proc/sys/net/ipv4/ip_forward",
		[]byte("1"),
		0644,
	); err != nil {
		return errors.New("failed to enable IP forwarding: " + err.Error())
	}

	iptablesSetup()

	return nil
}
