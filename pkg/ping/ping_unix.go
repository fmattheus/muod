//go:build !windows
package ping

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type unixPinger struct {
	conn *icmp.PacketConn
}

func newPinger() (Pinger, error) {
	conn, err := icmp.ListenPacket("udp4", "")
	if err != nil {
		return nil, err
	}
	return &unixPinger{conn: conn}, nil
}

func (up *unixPinger) Close() error {
	if up.conn != nil {
		return up.conn.Close()
	}
	return nil
}

func createICMPMessage(id, seq int) []byte {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: []byte("ping"),
		},
	}
	
	msgBytes, _ := msg.Marshal(nil)
	return msgBytes
}

func (up *unixPinger) Ping(ip net.IP, timeout time.Duration) (time.Duration, error) {
	if err := up.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}

	msg := createICMPMessage(os.Getpid()&0xffff, 1)
	if _, err := up.conn.WriteTo(msg, &net.UDPAddr{IP: ip}); err != nil {
		return 0, err
	}

	start := time.Now()

	reply := make([]byte, 1500)
	n, _, err := up.conn.ReadFrom(reply)
	if err != nil {
		return 0, err
	}

	rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])
	if err != nil {
		return 0, err
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return time.Since(start), nil
	default:
		return 0, fmt.Errorf("unexpected ICMP message type: %v", rm.Type)
	}
} 