// Package ping provides a cross-platform implementation of ICMP echo request (ping)
// functionality that works without requiring root/administrator privileges.
//
// On Unix-like systems (Linux, macOS, BSD), it uses unprivileged UDP sockets.
// On Windows, it uses the Windows ICMP Helper API (iphlpapi.dll).
//
// Example usage:
//
//	pinger, err := ping.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer pinger.Close()
//
//	rtt, err := pinger.Ping(ip, timeout)
//	if err != nil {
//	    log.Printf("Ping failed: %v", err)
//	} else {
//	    log.Printf("Host is up, RTT: %v", rtt)
//	}
package ping

import (
	"fmt"
	"net"
	"time"
)

// Result represents the result of a ping attempt
type Result struct {
	Host    string        // The hostname or IP address that was pinged
	Success bool          // Whether the ping was successful
	RTT     time.Duration // Round-trip time if successful
	Error   error        // Error message if unsuccessful
}

// HostInfo represents a resolved host with its IPv4 address
type HostInfo struct {
	Hostname string  // The original hostname provided
	IPAddr   net.IP  // The resolved IPv4 address
}

// Pinger defines the interface for platform-specific ping implementations.
// Each platform (Unix-like systems and Windows) provides its own implementation
// of this interface.
type Pinger interface {
	// Ping sends an ICMP echo request to the specified IP address and waits
	// for a response up to the specified timeout duration. It returns the
	// round-trip time if successful, or an error if the ping failed.
	Ping(net.IP, time.Duration) (time.Duration, error)
	
	// Close releases any resources used by the Pinger.
	// This method should always be called when done with the Pinger.
	Close() error
}

// ResolveHosts converts a list of hostnames to their corresponding IPv4 addresses.
// It returns a slice of HostInfo containing both the original hostname and its
// resolved IPv4 address. If any hostname cannot be resolved or does not have
// an IPv4 address, an error is returned.
func ResolveHosts(hosts []string) ([]HostInfo, error) {
	resolved := make([]HostInfo, 0, len(hosts))
	
	for _, host := range hosts {
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %v", host, err)
		}

		var ipv4Addr net.IP
		for _, ip := range ips {
			if ip.To4() != nil {
				ipv4Addr = ip
				break
			}
		}

		if ipv4Addr == nil {
			return nil, fmt.Errorf("no IPv4 address found for %s", host)
		}

		resolved = append(resolved, HostInfo{Hostname: host, IPAddr: ipv4Addr})
	}
	
	return resolved, nil
}

// New creates a new platform-specific Pinger implementation.
// On Unix-like systems, it creates a UDP-based pinger.
// On Windows, it creates a pinger using the ICMP Helper API.
func New() (Pinger, error) {
	return newPinger()
} 