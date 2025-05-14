//go:build !windows
package ping

import (
	"net"
	"testing"
	"time"
)

// TestUnixPingerCreation tests Unix-specific pinger creation details
func TestUnixPingerCreation(t *testing.T) {
	p, err := newUnixPinger()
	if err != nil {
		t.Fatalf("Failed to create Unix pinger: %v", err)
	}
	defer p.Close()

	// Check Unix-specific implementation details
	if p.conn == nil {
		t.Error("Expected non-nil connection in Unix pinger")
	}
}

// TestUnixSocketTimeout tests Unix-specific socket timeout handling
func TestUnixSocketTimeout(t *testing.T) {
	pinger, err := newUnixPinger()
	if err != nil {
		t.Fatalf("Failed to create Unix pinger: %v", err)
	}
	defer pinger.Close()

	// Set a very short timeout
	timeout := 1 * time.Millisecond
	ip := net.ParseIP("8.8.8.8") // Use Google DNS, but timeout will occur

	_, err = pinger.Ping(ip, timeout)
	if err == nil {
		t.Error("Expected timeout error for short timeout")
	}
}

// TestUnixICMPMessageCreation tests ICMP message creation
func TestUnixICMPMessageCreation(t *testing.T) {
	msg := createICMPMessage(1234, 5678)
	if len(msg) == 0 {
		t.Error("Expected non-empty ICMP message")
	}
}

// TestUnixMultipleClose tests multiple Close() calls
func TestUnixMultipleClose(t *testing.T) {
	pinger, err := newUnixPinger()
	if err != nil {
		t.Fatalf("Failed to create Unix pinger: %v", err)
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