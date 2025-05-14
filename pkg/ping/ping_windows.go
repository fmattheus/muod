//go:build windows
package ping

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows ICMP API constants
const (
	ICMP_SUCCESS      = 0
	IP_HEADER_LENGTH  = 20
	ICMP_ECHO_REQUEST = 8
)

// IP_OPTION_INFORMATION structure
type ipOptionInformation struct {
	TTL         uint8
	TOS         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

// ICMP_ECHO_REPLY structure
type icmpEchoReply struct {
	Address       [4]byte
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       ipOptionInformation
}

type windowsPinger struct {
	handle windows.Handle
	dll    *windows.DLL
	proc   *windows.Proc
}

func newPinger() (Pinger, error) {
	dll, err := windows.LoadDLL("iphlpapi.dll")
	if err != nil {
		return nil, fmt.Errorf("failed to load iphlpapi.dll: %v", err)
	}

	proc, err := dll.FindProc("IcmpCreateFile")
	if err != nil {
		dll.Release()
		return nil, fmt.Errorf("failed to find IcmpCreateFile: %v", err)
	}

	handle, _, err := proc.Call()
	if handle == 0 {
		dll.Release()
		return nil, fmt.Errorf("IcmpCreateFile failed: %v", err)
	}

	return &windowsPinger{
		handle: windows.Handle(handle),
		dll:    dll,
		proc:   proc,
	}, nil
}

func (wp *windowsPinger) Close() error {
	if wp.handle != 0 {
		closeProc, err := wp.dll.FindProc("IcmpCloseHandle")
		if err == nil {
			closeProc.Call(uintptr(wp.handle))
		}
	}
	return wp.dll.Release()
}

func (wp *windowsPinger) Ping(ip net.IP, timeout time.Duration) (time.Duration, error) {
	sendProc, err := wp.dll.FindProc("IcmpSendEcho")
	if err != nil {
		return 0, fmt.Errorf("failed to find IcmpSendEcho: %v", err)
	}

	timeoutMs := uint32(timeout.Milliseconds())
	if timeoutMs < 1 {
		timeoutMs = 1
	}

	data := []byte("ping")
	replySize := uint32(unsafe.Sizeof(icmpEchoReply{})) + uint32(len(data))
	replyBuf := make([]byte, replySize)

	ipAddr := binary.LittleEndian.Uint32(ip.To4())

	ret, _, err := sendProc.Call(
		uintptr(wp.handle),
		uintptr(ipAddr),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		0,
		uintptr(unsafe.Pointer(&replyBuf[0])),
		uintptr(replySize),
		uintptr(timeoutMs),
	)

	if ret == 0 {
		return 0, fmt.Errorf("IcmpSendEcho failed: %v", err)
	}

	reply := (*icmpEchoReply)(unsafe.Pointer(&replyBuf[0]))
	return time.Duration(reply.RoundTripTime) * time.Millisecond, nil
} 