//go:build all || cpdp

package static

import (
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestUdpStaticLag(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":   uint64(50),
		"pktCount":  uint32(100),
		"pktSize":   uint32(128),
		"txMac":     "00:00:01:01:01:01",
		"txIp":      "1.1.1.1",
		"rxMac":     "00:00:01:01:01:02",
		"rxIp":      "1.1.1.2",
		"txUdpPort": uint32(5000),
		"rxUdpPort": uint32(6000),
	}

	api := otg.NewOtgApi(t)
	c := udpStaticLagConfig(api, testConst)

	api.SetConfig(c)

	api.StartCapture()
	// TODO: do we need to wait after starting protocols ?
	api.StartProtocols()
	// TODO: do we need to check LAG metrics ?
	api.StartTransmit()

	api.WaitFor(
		func() bool { return udpStaticLagFlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopCapture()

	// TODO: check capture on correct port
	// udpStaticLagCaptureOk(api, c, testConst)
}

func udpStaticLagConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(api.TestConfig().OtgPorts[0])
	p2 := c.Ports().Add().SetName("p2").SetLocation(api.TestConfig().OtgPorts[1])
	p3 := c.Ports().Add().SetName("p3").SetLocation(api.TestConfig().OtgPorts[2])
	p4 := c.Ports().Add().SetName("p4").SetLocation(api.TestConfig().OtgPorts[3])
	p5 := c.Ports().Add().SetName("p5").SetLocation(api.TestConfig().OtgPorts[4])
	p6 := c.Ports().Add().SetName("p6").SetLocation(api.TestConfig().OtgPorts[5])

	l1 := c.Lags().Add().SetName("l1").SetMinLinks(2)
	l1.Protocol().Static().SetLagId(1)
	l1p1 := l1.Ports().Add().SetPortName(p1.Name())
	l1p1.Ethernet().
		SetMac("00:00:00:00:00:01").
		SetName(l1.Name() + p1.Name())
	l1p2 := l1.Ports().Add().SetPortName(p2.Name())
	l1p2.Ethernet().
		SetMac("00:00:00:00:00:02").
		SetName(l1.Name() + p2.Name())
	l1p3 := l1.Ports().Add().SetPortName(p3.Name())
	l1p3.Ethernet().
		SetMac("00:00:00:00:00:03").
		SetName(l1.Name() + p3.Name())

	l2 := c.Lags().Add().SetName("l2").SetMinLinks(2)
	l2.Protocol().Static().SetLagId(1)
	l2p4 := l2.Ports().Add().SetPortName(p4.Name())
	l2p4.Ethernet().
		SetMac("00:00:00:00:00:04").
		SetName(l2.Name() + p4.Name())
	l2p5 := l2.Ports().Add().SetPortName(p5.Name())
	l2p5.Ethernet().
		SetMac("00:00:00:00:00:05").
		SetName(l2.Name() + p5.Name())
	l2p6 := l2.Ports().Add().SetPortName(p6.Name())
	l2p6.Ethernet().
		SetMac("00:00:00:00:00:06").
		SetName(l2.Name() + p6.Name())

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name(), p2.Name(), p3.Name(), p4.Name(), p5.Name(), p6.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	if api.TestConfig().OtgCaptureCheck {
		c.Captures().Add().
			SetName("ca").
			SetPortNames([]string{p1.Name(), p4.Name()}).
			SetFormat(gosnappi.CaptureFormat.PCAP)
	}

	f1 := c.Flows().Add().SetName("f1")
	f1.TxRx().Port().
		SetTxName(l1.Name()).
		SetRxName(l2.Name())

	f1.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
	f1.Rate().SetPps(tc["pktRate"].(uint64))
	f1.Size().SetFixed(tc["pktSize"].(uint32))
	f1.Metrics().SetEnable(true)

	eth := f1.Packet().Add().Ethernet()
	eth.Src().SetValue(tc["txMac"].(string))
	eth.Dst().SetValue(tc["rxMac"].(string))

	ip := f1.Packet().Add().Ipv4()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	udp := f1.Packet().Add().Udp()
	udp.SrcPort().SetValue(tc["txUdpPort"].(uint32))
	udp.DstPort().SetValue(tc["rxUdpPort"].(uint32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func udpStaticLagFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	m := api.GetFlowMetrics()[0]
	expCount := uint64(tc["pktCount"].(uint32))

	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == expCount &&
		m.FramesRx() == expCount
}

func udpStaticLagCaptureOk(api *otg.OtgApi, c gosnappi.Config, tc map[string]interface{}) {
	if !api.TestConfig().OtgCaptureCheck {
		return
	}
	ignoredCount := 0
	cPackets := api.GetCapture(c.Ports().Items()[3].Name())
	t := api.Testing()

	for i := 0; i < len(cPackets.Packets); i++ {
		// ignore unexpected packets based on ethernet src MAC
		if !cPackets.HasField(t, "ethernet src", i, 6, api.MacAddrToBytes(tc["txMac"].(string))) {
			ignoredCount += 1
			continue
		}
		// packet size
		cPackets.ValidateSize(t, i, int(tc["pktSize"].(uint32)))
		// ethernet header
		cPackets.ValidateField(t, "ethernet dst", i, 0, api.MacAddrToBytes(tc["rxMac"].(string)))
		cPackets.ValidateField(t, "ethernet type", i, 12, api.Uint64ToBytes(2048, 2))
		// ipv4 header
		cPackets.ValidateField(t, "ipv4 total length", i, 16, api.Uint64ToBytes(uint64(tc["pktSize"].(uint32)-14-4), 2))
		cPackets.ValidateField(t, "ipv4 protocol", i, 23, api.Uint64ToBytes(17, 1))
		cPackets.ValidateField(t, "ipv4 src", i, 26, api.Ipv4AddrToBytes(tc["txIp"].(string)))
		cPackets.ValidateField(t, "ipv4 dst", i, 30, api.Ipv4AddrToBytes(tc["rxIp"].(string)))
		// udp header
		cPackets.ValidateField(t, "udp src", i, 34, api.Uint64ToBytes(uint64(tc["txUdpPort"].(uint32)), 2))
		cPackets.ValidateField(t, "udp dst", i, 36, api.Uint64ToBytes(uint64(tc["rxUdpPort"].(uint32)), 2))
		cPackets.ValidateField(t, "udp length", i, 38, api.Uint64ToBytes(uint64(tc["pktSize"].(uint32)-14-4-20), 2))
	}

	expCount := int(tc["pktCount"].(uint32))
	actCount := len(cPackets.Packets) - ignoredCount
	if expCount != actCount {
		t.Fatalf("ERROR: expCount %d != actCount %d\n", expCount, actCount)
	}
}
