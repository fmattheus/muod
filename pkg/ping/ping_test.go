package ping

import (
	"net"
	"runtime"
	"testing"
	"time"
)

// pinger defines the interface that both Unix and Windows implementations must satisfy
type pinger interface {
	Ping(net.IP, time.Duration) (time.Duration, error)
	Close() error
}

// newPinger creates the appropriate pinger for the current OS
func newPinger(t *testing.T) pinger {
	var p pinger
	var err error
	
	if runtime.GOOS == "windows" {
		p, err = newWindowsPinger()
	} else {
		p, err = newUnixPinger()
	}
	if err != nil {
		t.Fatalf("Failed to create pinger: %v", err)
	}
	return p
}

// testHosts contains a mix of reliable and unreliable hosts for testing
var testHosts = []struct {
	name     string
	ip       string
	expected bool // true if we expect this host to respond
}{
	{"localhost", "127.0.0.1", true},
	{"google-dns", "8.8.8.8", true},
	{"invalid", "0.0.0.0", false},
}

// TestHostResolution tests the host resolution functionality
func TestHostResolution(t *testing.T) {
	hosts := []string{"localhost", "google.com"}
	resolved, err := resolveHosts(hosts)
	if err != nil {
		t.Fatalf("Failed to resolve hosts: %v", err)
	}

	if len(resolved) != len(hosts) {
		t.Errorf("Expected %d resolved hosts, got %d", len(hosts), len(resolved))
	}

	// Check localhost resolution
	found := false
	for _, host := range resolved {
		if host.hostname == "localhost" {
			if !host.ipAddr.Equal(net.ParseIP("127.0.0.1")) {
				t.Errorf("Expected localhost to resolve to 127.0.0.1, got %v", host.ipAddr)
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("Failed to find localhost in resolved hosts")
	}
}

// TestPingTimeout tests that pings timeout appropriately
func TestPingTimeout(t *testing.T) {
	p := newPinger(t)
	defer p.Close()

	// Test with very short timeout to unreachable host
	unreachableIP := net.ParseIP("192.0.2.1") // TEST-NET-1 from RFC 5737
	timeout := 100 * time.Millisecond
	
	_, err := p.Ping(unreachableIP, timeout)
	if err == nil {
		t.Error("Expected timeout error for unreachable host")
	}
}

// TestPingValidHost tests pinging a known good host
func TestPingValidHost(t *testing.T) {
	p := newPinger(t)
	defer p.Close()

	// Test localhost
	ip := net.ParseIP("127.0.0.1")
	rtt, err := p.Ping(ip, time.Second)
	if err != nil {
		t.Errorf("Failed to ping localhost: %v", err)
	}
	if rtt <= 0 {
		t.Error("Expected positive RTT for localhost")
	}
}

// TestMultipleHosts tests pinging multiple hosts in sequence
func TestMultipleHosts(t *testing.T) {
	hosts := []string{"localhost", "127.0.0.1"}
	resolved, err := resolveHosts(hosts)
	if err != nil {
		t.Fatalf("Failed to resolve hosts: %v", err)
	}

	results := pingHosts(resolved)
	if len(results) != len(hosts) {
		t.Errorf("Expected %d results, got %d", len(hosts), len(results))
	}

	// At least one of the localhost pings should succeed
	success := false
	for _, result := range results {
		if result.success {
			success = true
			break
		}
	}
	if !success {
		t.Error("Expected at least one successful ping to localhost")
	}
}

// TestMultipleClose tests multiple Close() calls
func TestMultipleClose(t *testing.T) {
	p := newPinger(t)

	// First close should succeed
	if err := p.Close(); err != nil {
		t.Errorf("First close failed: %v", err)
	}

	// Second close should not error
	if err := p.Close(); err != nil {
		t.Errorf("Second close failed: %v", err)
	}
} 