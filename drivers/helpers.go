package drivers

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// Returns this computer's address on the interface
func GetInterfaceIPv4Addr(iface string) (addr net.IP, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)

	// Get interface by name
	if ief, err = net.InterfaceByName(iface); err != nil { // get interface
		return
	}

	// Find the first IPv4 address on the NIC
	if addrs, err = ief.Addrs(); err != nil {
		return
	}

	for _, addr := range addrs {
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}

	// If no IPv4 addresses were found on the interface
	if ipv4Addr == nil {
		return nil, fmt.Errorf("interface %s don't have an ipv4 address", iface)
	}

	return ipv4Addr, nil
}

func GetTransportForIPv4(ifaceIP net.IP) (*http.Transport, error) {
	addr, err := net.ResolveTCPAddr("tcp", ifaceIP.String()+":0")
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{LocalAddr: addr}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dialer.Dial(network, addr)
		return conn, err
	}

	return &http.Transport{DialContext: dialContext}, nil
}
