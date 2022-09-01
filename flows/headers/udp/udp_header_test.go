//go:build all || free || b2b

package udp

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestUdpHeader(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":   int64(50),
		"pktCount":  int32(100),
		"pktSize":   int32(128),
		"txMac":     "00:00:01:01:01:01",
		"rxMac":     "00:00:01:01:01:02",
		"txIp":      "1.1.1.1",
		"rxIp":      "1.1.1.2",
		"txUdpPort": int32(5000),
		"rxUdpPort": int32(6000),
	}

	api := otg.NewOtgApi(t)
	c := udpHeaderConfig(api, testConst)

	api.SetConfig(c)

	api.StartCapture()
	api.StartTransmit()

	api.WaitFor(
		func() bool { return udpHeaderMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopCapture()

	udpHeaderCaptureOk(api, c, testConst)
}

func udpHeaderConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(api.TestConfig().OtgPorts[0])
	p2 := c.Ports().Add().SetName("p2").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name(), p2.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	c.Captures().Add().
		SetName("ca").
		SetPortNames([]string{p1.Name(), p2.Name()}).
		SetFormat(gosnappi.CaptureFormat.PCAP)

	f1 := c.Flows().Add().SetName("f1")
	f1.TxRx().Port().
		SetTxName(p1.Name()).
		SetRxName(p2.Name())
	f1.Duration().FixedPackets().SetPackets(tc["pktCount"].(int32))
	f1.Rate().SetPps(tc["pktRate"].(int64))
	f1.Size().SetFixed(tc["pktSize"].(int32))
	f1.Metrics().SetEnable(true)

	eth := f1.Packet().Add().Ethernet()
	eth.Src().SetValue(tc["txMac"].(string))
	eth.Dst().SetValue(tc["rxMac"].(string))

	ip := f1.Packet().Add().Ipv4()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	udp := f1.Packet().Add().Udp()
	udp.SrcPort().SetValue(tc["txUdpPort"].(int32))
	udp.DstPort().SetValue(tc["rxUdpPort"].(int32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func udpHeaderMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	m := api.GetFlowMetrics()[0]
	expCount := int64(tc["pktCount"].(int32))

	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == expCount &&
		m.FramesRx() == expCount
}

func udpHeaderCaptureOk(api *otg.OtgApi, c gosnappi.Config, tc map[string]interface{}) {
	if !api.TestConfig().OtgCaptureCheck {
		return
	}
	expCount := int(tc["pktCount"].(int32))
	cPackets := api.GetCapture(c.Ports().Items()[1].Name())

	for i := 0; i < expCount; i++ {
		// ethernet header
		if err := cPackets.ValidateField(i, 0, api.MacAddrToBytes(tc["rxMac"].(string))); err != nil {
			api.Testing().Fatalf("ethernet rxMac not ok: %v\n", err)
		}
		if err := cPackets.ValidateField(i, 6, api.MacAddrToBytes(tc["txMac"].(string))); err != nil {
			api.Testing().Fatalf("ethernet txMac not ok: %v\n", err)
		}
		if err := cPackets.ValidateField(i, 12, api.Uint64ToBytes(2048, 2)); err != nil {
			api.Testing().Fatalf("ethernet type not ok: %v\n", err)
		}
		// ipv4 header
		if err := cPackets.ValidateField(i, 16, api.Uint64ToBytes(uint64(tc["pktSize"].(int32)-14-4), 2)); err != nil {
			api.Testing().Fatalf("ipv4 totalLength not ok: %v\n", err)
		}
		if err := cPackets.ValidateField(i, 23, api.Uint64ToBytes(17, 1)); err != nil {
			api.Testing().Fatalf("ipv4 protocol not ok: %v\n", err)
		}
		if err := cPackets.ValidateField(i, 26, api.Ipv4AddrToBytes(tc["txIp"].(string))); err != nil {
			api.Testing().Fatalf("ipv4 src not ok: %v\n", err)
		}
		if err := cPackets.ValidateField(i, 30, api.Ipv4AddrToBytes(tc["rxIp"].(string))); err != nil {
			api.Testing().Fatalf("ipv4 dst not ok: %v\n", err)
		}
		// udp header
		if err := cPackets.ValidateField(i, 34, api.Uint64ToBytes(uint64(tc["txUdpPort"].(int32)), 2)); err != nil {
			api.Testing().Fatalf("udp src not ok: %v\n", err)
		}
		if err := cPackets.ValidateField(i, 36, api.Uint64ToBytes(uint64(tc["rxUdpPort"].(int32)), 2)); err != nil {
			api.Testing().Fatalf("udp dst not ok: %v\n", err)
		}
		if err := cPackets.ValidateField(i, 38, api.Uint64ToBytes(uint64(tc["pktSize"].(int32)-14-4-20), 2)); err != nil {
			api.Testing().Fatalf("udp length not ok: %v\n", err)
		}
	}
}
