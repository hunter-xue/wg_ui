package wg

import (
	"fmt"
	"net"
	"strings"
)

// ipFromCIDR extracts the IP part from a CIDR string (e.g. "10.0.0.1/32" → "10.0.0.1").
func ipFromCIDR(cidr string) string {
	if idx := strings.Index(cidr, "/"); idx >= 0 {
		return cidr[:idx]
	}
	return cidr
}

// SameSubnet reports whether clientCIDR is in the same subnet as serverCIDR.
// WireGuard interface addresses are commonly /32; in that case a /24 supernet is used
// for the check, which covers the typical VPN addressing pattern.
func SameSubnet(serverCIDR, clientCIDR string) (bool, error) {
	serverIP, serverNet, err := net.ParseCIDR(serverCIDR)
	if err != nil {
		return false, fmt.Errorf("invalid server address %q: %w", serverCIDR, err)
	}

	clientIPStr := ipFromCIDR(clientCIDR)
	clientIP := net.ParseIP(clientIPStr)
	if clientIP == nil {
		return false, fmt.Errorf("invalid client address %q", clientIPStr)
	}

	ones, _ := serverNet.Mask.Size()
	subnet := serverNet
	if ones >= 31 {
		// /31 or /32 are host routes; derive a /24 supernet for the VPN range check.
		_, subnet, err = net.ParseCIDR(serverIP.String() + "/24")
		if err != nil {
			return false, err
		}
	}

	return subnet.Contains(clientIP), nil
}

// SameIP reports whether two CIDR addresses refer to the same host IP.
func SameIP(cidrA, cidrB string) bool {
	return ipFromCIDR(cidrA) == ipFromCIDR(cidrB)
}
