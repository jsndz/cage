package network

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

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
	netlink.LinkSetUp(link)
	lo, _ := netlink.LinkByName("lo")
	netlink.LinkSetUp(lo)
}

func SetUpContainerNetwork(containerPid int, bridge *netlink.Bridge, peernet, hostnet string) {
	CreateVethPair(hostnet, peernet)
	//connect host to bridge
	ConnectToBridge(hostnet, bridge)
	// move peer to container
	ConnectContainer(hostnet, peernet, containerPid)
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
