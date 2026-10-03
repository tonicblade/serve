package network

import (
	"net"
)

func DiscoverAddresses(host string) []string {
	if host != "0.0.0.0" && host != "::" && host != "" {
		return nil
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	seen := map[string]struct{}{}
	addrs := make([]string, 0, 4)
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, rawAddr := range ifaceAddrs {
			ipNet, ok := rawAddr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if _, ok := seen[ip.String()]; ok {
				continue
			}
			seen[ip.String()] = struct{}{}
			addrs = append(addrs, ip.String())
		}
	}
	if len(addrs) == 0 {
		addrs = append(addrs, "127.0.0.1")
	}
	return addrs
}
