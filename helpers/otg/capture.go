package otg

import (
	"io"
	"os"
	"time"

	"github.com/dreadl0ck/gopcap"
)

func (o *OtgApi) GetCapture(portName string) []byte {
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
		t.Fatalf("Could not create temporary pcap file: %v\n", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(res); err != nil {
		t.Fatalf("Could not write bytes to pcap file: %v\n", err)
	}
	f.Close()

	r, err := gopcap.Open(f.Name())
	if err != nil {
		t.Fatalf("Could not open pcap file %s: %v\n", f.Name(), err)
	}
	defer r.Close()

	for {
		h, data, err := r.ReadNextPacket()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Could not read next packet: %v\n", err)
		}
		t.Log(h, len(data))
	}
	return nil
}
