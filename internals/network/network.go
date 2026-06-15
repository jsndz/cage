package network

import (
	"fmt"
	"net"
	"os"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

var nextIP = 1

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
	// equivalent of
	//ip link add cage0 type bridge
	// ip link set cage0 up

	if err := netlink.LinkAdd(bridge); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(bridge); err != nil {
		return nil, err
	}
	AssignIP(bridge, fmt.Sprintf("10.0.0.%d/24", nextIP))
	nextIP++

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

func ConnectContainer(hostname, peername string, containerPID int) {
	// find the veth pair

	host, _ := netlink.LinkByName(hostname)
	netlink.LinkSetUp(host)
	peer, _ := netlink.LinkByName(peername)
	netlink.LinkSetNsPid(peer, containerPID)
}

func SetUpVeth(peername string) {
	link, _ := netlink.LinkByName(peername)
	AssignIP(link, fmt.Sprintf("10.0.0.%d/24", nextIP))
	netlink.LinkSetUp(link)

	lo, _ := netlink.LinkByName("lo")
	netlink.LinkSetUp(lo)
	// routing all traffic through the bridge
	// for example, if we want to access google.com from the container
	// the traffic will go from the container to the bridge and then to the internet through the host's network interface
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        net.ParseIP("10.0.0.1"),
	}

	netlink.RouteAdd(route)

}

func SetUpContainerNetwork(containerPid int, bridge *netlink.Bridge, peernet, hostnet string) {
	CreateVethPair(hostnet, peernet)
	//connect host to bridge
	ConnectToBridge(hostnet, bridge)
	// move peer to container
	ConnectContainer(hostnet, peernet, containerPid)

	NftableSetup()
}

func CleanBridge(hostnet string) error {
	link, err := netlink.LinkByName("")
	if err != nil {
		return err
	}
	err = netlink.LinkDel(link)
	if err != nil {
		return err
	}
	return nil
}

func getDefaultInterface() (string, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return "", err
	}

	for _, r := range routes {
		if r.Dst == nil { // default route
			link, err := netlink.LinkByIndex(r.LinkIndex)
			if err != nil {
				return "", err
			}

			return link.Attrs().Name, nil
		}
	}
	return "", fmt.Errorf("default interface not found")
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
		return err
	}

	natExists, err := tableExists(conn, nftables.TableFamilyIPv4, "cage_nat")
	if err != nil {
		return err
	}
	filterExists, err := tableExists(conn, nftables.TableFamilyINet, "cage_filter")
	if err != nil {
		return err
	}
	if natExists && filterExists {
		return nil
	}
	inf, err := getDefaultInterface()
	if err != nil {
		return err
	}

	natTable := conn.AddTable(&nftables.Table{
		Family: nftables.TableFamilyIPv4,
		Name:   "cage_nat",
	})
	// A chain in nftables is a container for rules.
	postrouting := conn.AddChain(&nftables.Chain{
		Name:     "postrouting",
		Table:    natTable,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})
	// rules represent how data will be handled
	// here we are matching packets with source IP in 10.0.0.0/24 and outgoing interface eth0, and then applying masquerading to them
	//i.e, data going from eth0 with ip 10.0.0.0/24 masquerade the ip with some other ip meaning giving public ip
	_, subnet, err := net.ParseCIDR("10.0.0.0/24")
	if err != nil {
		return err
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

			// Match "eth0"
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     ifnameBytes(inf),
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
	// container -> internet allowed if outgoing interface is eth0 and incoming interface is cage0
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

			// oifname "eth0"
			&expr.Meta{
				Key:      expr.MetaKeyOIFNAME,
				Register: 1,
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     ifnameBytes(inf),
			},

			// accept
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})
	// internet -> container allowed if incoming interface is eth0 and outgoing interface is cage0 and connection state is established or related
	conn.AddRule(&nftables.Rule{
		Table: filterTable,
		Chain: forwarding,
		Exprs: []expr.Any{
			// iifname "eth0"
			&expr.Meta{
				Key:      expr.MetaKeyIIFNAME,
				Register: 1,
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     ifnameBytes(inf),
			},

			// oifname "cage0"
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

			// accept
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})
	if err := conn.Flush(); err != nil {
		return err
	}
	if err := os.WriteFile(
		"/proc/sys/net/ipv4/ip_forward",
		[]byte("1"),
		0644,
	); err != nil {
		return err
	}
	return nil
}
