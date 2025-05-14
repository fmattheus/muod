//go:build windows
package ping

import (
	"net"
	"testing"
	"time"
)

// TestWindowsPingerCreation tests Windows-specific pinger creation details
func TestWindowsPingerCreation(t *testing.T) {
	p, err := newWindowsPinger()
	if err != nil {
		t.Fatalf("Failed to create Windows pinger: %v", err)
	}
	defer p.Close()

	// Check Windows-specific implementation details
	if p.handle == 0 {
		t.Error("Expected non-zero handle in Windows pinger")
	}
	if p.dll == nil {
		t.Error("Expected non-nil DLL in Windows pinger")
	}
}

// TestWindowsDLLHandling tests Windows-specific DLL handling
func TestWindowsDLLHandling(t *testing.T) {
	p, err := newWindowsPinger()
	if err != nil {
		t.Fatalf("Failed to create Windows pinger: %v", err)
	}
	defer p.Close()

	// Verify we can find the required procedures
	sendProc, err := p.dll.FindProc("IcmpSendEcho")
	if err != nil {
		t.Errorf("Failed to find IcmpSendEcho: %v", err)
	}
	if sendProc == nil {
		t.Error("Expected non-nil IcmpSendEcho procedure")
	}
}

// TestWindowsIPHLPAPI tests the Windows IP Helper API functionality
func TestWindowsIPHLPAPI(t *testing.T) {
	pinger, err := newWindowsPinger()
	if err != nil {
		t.Fatalf("Failed to create Windows pinger: %v", err)
	}
	defer pinger.Close()

	// Test localhost ping
	ip := net.ParseIP("127.0.0.1")
	rtt, err := pinger.Ping(ip, time.Second)
	if err != nil {
		t.Errorf("Failed to ping localhost: %v", err)
	}
	if rtt <= 0 {
		t.Error("Expected positive RTT for localhost")
	}
}

// TestWindowsInvalidIP tests handling of invalid IP addresses
func TestWindowsInvalidIP(t *testing.T) {
	pinger, err := newWindowsPinger()
	if err != nil {
		t.Fatalf("Failed to create Windows pinger: %v", err)
	}
	defer pinger.Close()

	// Test with invalid IP
	ip := net.ParseIP("0.0.0.0")
	_, err = pinger.Ping(ip, time.Second)
	if err == nil {
		t.Error("Expected error when pinging invalid IP")
	}
}

// TestWindowsMultipleClose tests multiple Close() calls
func TestWindowsMultipleClose(t *testing.T) {
	pinger, err := newWindowsPinger()
	if err != nil {
		t.Fatalf("Failed to create Windows pinger: %v", err)
	}

	// First close should succeed
	if err := pinger.Close(); err != nil {
		t.Errorf("First close failed: %v", err)
	}

	// Second close should not error
	if err := pinger.Close(); err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}

// TestWindowsTimeoutHandling tests timeout behavior
func TestWindowsTimeoutHandling(t *testing.T) {
	pinger, err := newWindowsPinger()
	if err != nil {
		t.Fatalf("Failed to create Windows pinger: %v", err)
	}
	defer pinger.Close()

	// Use TEST-NET-1 (RFC 5737) for timeout test
	ip := net.ParseIP("192.0.2.1")
	timeout := 100 * time.Millisecond

	_, err = pinger.Ping(ip, timeout)
	if err == nil {
		t.Error("Expected timeout error for unreachable host")
	}
} 