//go:build all || dp

package udp

import (
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestUdpPortValue(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":   uint64(50),
		"pktCount":  uint32(100),
		"pktSize":   uint32(128),
		"txMac":     "00:00:01:01:01:01",
		"rxMac":     "00:00:01:01:01:02",
		"txIp":      "1.1.1.1",
		"rxIp":      "1.1.1.2",
		"txUdpPort": uint32(5000),
		"rxUdpPort": uint32(6000),
	}

	api := otg.NewOtgApi(t)
	c := udpPortValueConfig(api, testConst)

	api.SetConfig(c)

	api.StartCapture()
	api.StartTransmit()

	api.WaitFor(
		func() bool { return udpPortValueFlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopCapture()

	udpPortValueCaptureOk(api, c, testConst)
}

func udpPortValueConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(api.TestConfig().OtgPorts[0])
	p2 := c.Ports().Add().SetName("p2").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name(), p2.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	if api.TestConfig().OtgCaptureCheck {
		c.Captures().Add().
			SetName("ca").
			SetPortNames([]string{p1.Name(), p2.Name()}).
			SetFormat(gosnappi.CaptureFormat.PCAP)
	}

	f1 := c.Flows().Add().SetName("f1")
	f1.TxRx().Port().
		SetTxName(p1.Name()).
		SetRxName(p2.Name())
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

func udpPortValueFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	m := api.GetFlowMetrics()[0]
	expCount := uint64(tc["pktCount"].(uint32))

	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == expCount &&
		m.FramesRx() == expCount
}

func udpPortValueCaptureOk(api *otg.OtgApi, c gosnappi.Config, tc map[string]interface{}) {
	if !api.TestConfig().OtgCaptureCheck {
		return
	}
	ignoredCount := 0
	cPackets := api.GetCapture(c.Ports().Items()[1].Name())
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
