package otg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/dreadl0ck/gopcap"
)

type CapturedPacket struct {
	Sequence       int
	Timestamp      time.Time
	Data           []byte
	ActualLength   int
	CapturedLength int
}

type CapturedPackets struct {
	Packets []CapturedPacket
}

func (c *CapturedPackets) CheckField(sequence int, startOffSet int, field []byte) error {
	if sequence >= len(c.Packets) {
		return fmt.Errorf("sequence %d >= len(capturedPackets) %d", sequence, len(c.Packets))
	}

	p := c.Packets[sequence]
	if startOffSet < 0 || startOffSet >= len(p.Data) {
		return fmt.Errorf("startOffSet %d not in range [0, %d); data: %v", sequence, len(p.Data), p.Data)
	}

	endOffset := startOffSet + len(field) - 1
	if endOffset < startOffSet {
		return fmt.Errorf("startOffSet %d > endOffset %d; field: %v", startOffSet, endOffset, field)
	}

	if endOffset >= len(p.Data) {
		return fmt.Errorf("endOffset %d not in range [0, %d); field: %v data: %v", endOffset, len(p.Data), field, p.Data)
	}

	if !bytes.Equal(field, p.Data[startOffSet:endOffset+1]) {
		return fmt.Errorf("field %v != actualField %v; sequence: %d, startOffset: %d, data: %v", field, p.Data[startOffSet:endOffset+1], sequence, startOffSet, p.Data)
	}

	return nil
}

func (c *CapturedPackets) CheckSize(sequence int, size int) error {
	if len(c.Packets[sequence].Data) != size {
		return fmt.Errorf("expSize %d != actSize %d", size, len(c.Packets[sequence].Data))
	}

	return nil
}

func (c *CapturedPackets) ValidateSize(t *testing.T, sequence int, size int) {
	if err := c.CheckSize(sequence, size); err != nil {
		t.Fatalf("ERROR: %v\n", err)
	}
}

func (c *CapturedPackets) ValidateField(t *testing.T, name string, sequence int, startOffSet int, field []byte) {
	if err := c.CheckField(sequence, startOffSet, field); err != nil {
		t.Fatalf("ERROR: %s: %v\n", name, err)
	}
}

func (c *CapturedPackets) HasField(t *testing.T, name string, sequence int, startOffSet int, field []byte) bool {
	if err := c.CheckField(sequence, startOffSet, field); err != nil {
		t.Logf("WARNING: %s: %v\n", name, err)
		return false
	}

	return true
}

func (o *OtgApi) GetCapture(portName string) *CapturedPackets {
	t := o.Testing()
	api := o.Api()

	if !o.TestConfig().OtgCaptureCheck {
		t.Log("Skipped GetCapture")
		return nil
	}

	t.Logf("Getting capture from port %s ...\n", portName)
	defer o.Timer(time.Now(), "GetCapture")

	res, err := api.GetCapture(api.NewCaptureRequest().SetPortName(portName))
	o.LogWrnErr(nil, err, true)

	f, err := os.CreateTemp(".", "pcap")
	if err != nil {
		t.Fatalf("ERROR: Could not create temporary pcap file: %v\n", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(res); err != nil {
		t.Fatalf("ERROR: Could not write bytes to pcap file: %v\n", err)
	}
	f.Close()

	r, err := gopcap.Open(f.Name())
	if err != nil {
		t.Fatalf("ERROR: Could not open pcap file %s: %v\n", f.Name(), err)
	}
	defer r.Close()

	cPackets := CapturedPackets{
		Packets: []CapturedPacket{},
	}
	for i := 0; ; i++ {
		h, data, err := r.ReadNextPacket()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("ERROR: Could not read next packet: %v\n", err)
		}
		p := CapturedPacket{
			Sequence:       i,
			Data:           data,
			ActualLength:   int(h.OriginalLen),
			CapturedLength: int(h.CaptureLen),
			Timestamp:      time.UnixMicro(int64(h.TsSec)*1000000 + int64(h.TsUsec)),
		}
		cPackets.Packets = append(cPackets.Packets, p)
	}

	return &cPackets
}

func (o *OtgApi) MacAddrToBytes(mac string) []byte {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		o.Testing().Fatalf("ERROR: Could not parse MacAddr %s: %v\n", mac, err)
	}

	return hw
}

func (o *OtgApi) Ipv4AddrToBytes(ip string) []byte {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		o.Testing().Fatalf("ERROR: Could not parse IPv4Addr %s\n", ip)
	}

	v4 := parsedIp.To4()
	if v4 == nil {
		o.Testing().Fatalf("ERROR: Could not parse IPv4Addr %s\n", ip)
	}

	return v4
}

func (o *OtgApi) Ipv6AddrToBytes(ip string) []byte {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		o.Testing().Fatalf("ERROR: Could not parse IPv6Addr %s\n", ip)
	}

	v6 := parsedIp.To16()
	if v6 == nil {
		o.Testing().Fatalf("ERROR: Could not parse IPv6Addr %s\n", ip)
	}

	return v6
}

func (o *OtgApi) Uint64ToBytes(num uint64, size int) []byte {
	var b []byte

	if size < 8 {
		b = make([]byte, 8)
	} else {
		b = make([]byte, size)
	}

	binary.BigEndian.PutUint64(b, num)

	if size < 8 {
		return b[8-size : 8]
	}

	return b
}
