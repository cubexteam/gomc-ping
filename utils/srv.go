package utils

import (
	"fmt"
	"net"
	"strings"
)

// ResolveSRV looks up Minecraft SRV records and trims the trailing dot
func ResolveSRV(host string) (string, uint16, error) {
	_, addrs, err := net.LookupSRV("minecraft", "tcp", host)
	if err != nil {
		return "", 0, err
	}
	if len(addrs) == 0 {
		return "", 0, fmt.Errorf("no SRV records found")
	}

	// Trim the trailing dot from the target host
	target := strings.TrimSuffix(addrs[0].Target, ".")
	return target, addrs[0].Port, nil
}
