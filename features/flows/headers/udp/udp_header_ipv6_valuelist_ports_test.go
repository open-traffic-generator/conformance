//go:build all || feature || b2b || free_feature

package udp

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestUdpHeaderIpv6ValuelistPorts(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":            int64(50),
		"pktCount":           int32(100),
		"pktSize":            int32(128),
		"txMac":              "00:00:01:01:01:01",
		"rxMac":              "00:00:01:01:01:02",
		"txIp":               "2000::1",
		"rxIp":               "2000::2",
		"txUdpPortValueList": []int32{5000, 5010, 5020, 5030},
		"rxUdpPortValueList": []int32{6000, 6010, 6020, 6030},
	}

	api := otg.NewOtgApi(t)
	c := udpHeaderIpv6ValuelistPortsConfig(api, testConst)

	api.SetConfig(c)

	api.StartCapture()

	api.StartTransmit()

	api.WaitFor(
		func() bool { return udpHeaderIpv6ValuelistPortsMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopCapture()

	udpHeaderIpv6ValuelistPortsCaptureOk(api, c, testConst)
}

func udpHeaderIpv6ValuelistPortsConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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
	f1.Duration().FixedPackets().SetPackets(tc["pktCount"].(int32))
	f1.Rate().SetPps(tc["pktRate"].(int64))
	f1.Size().SetFixed(tc["pktSize"].(int32))
	f1.Metrics().SetEnable(true)

	eth := f1.Packet().Add().Ethernet()
	eth.Src().SetValue(tc["txMac"].(string))
	eth.Dst().SetValue(tc["rxMac"].(string))

	ip := f1.Packet().Add().Ipv6()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	udp := f1.Packet().Add().Udp()
	udp.SrcPort().SetValues(tc["txUdpPortValueList"].([]int32))
	udp.DstPort().SetValues(tc["rxUdpPortValueList"].([]int32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func udpHeaderIpv6ValuelistPortsMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	m := api.GetFlowMetrics()[0]
	expCount := int64(tc["pktCount"].(int32))

	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == expCount &&
		m.FramesRx() == expCount
}

func udpHeaderIpv6ValuelistPortsCaptureOk(api *otg.OtgApi, c gosnappi.Config, tc map[string]interface{}) {
	if !api.TestConfig().OtgCaptureCheck {
		return
	}
	ignoredCount := 0
	reqPacketidx := -1
	cPackets := api.GetCapture(c.Ports().Items()[1].Name())
	t := api.Testing()

	for i := 0; i < len(cPackets.Packets); i++ {
		// ignore unexpected packets based on ethernet src MAC
		if !cPackets.HasField(t, "ethernet src", i, 6, api.MacAddrToBytes(tc["txMac"].(string))) {
			ignoredCount += 1
			continue
		}
		reqPacketidx += 1
		// packet size
		if int32(len(cPackets.Packets[i].Data)) != tc["pktSize"].(int32) {
			t.Fatalf("ERROR: Expected Packet Size %d != Actual Packet Size %d\n", int32(len(cPackets.Packets[i].Data)), tc["pktSize"].(int32))
		}
		// ethernet header
		cPackets.ValidateField(t, "ethernet dst", i, 0, api.MacAddrToBytes(tc["rxMac"].(string)))
		cPackets.ValidateField(t, "ethernet src", i, 6, api.MacAddrToBytes(tc["txMac"].(string)))
		cPackets.ValidateField(t, "ethernet type", i, 12, api.Uint64ToBytes(34525, 2))
		// ipv6 header
		cPackets.ValidateField(t, "ipv6 src", i, 22, api.Ipv6AddrToBytes(tc["txIp"].(string)))
		cPackets.ValidateField(t, "ipv6 dst", i, 38, api.Ipv6AddrToBytes(tc["rxIp"].(string)))
		cPackets.ValidateField(t, "ipv6 next header", i, 20, api.Uint64ToBytes(17, 1))
		// udp header
		cPackets.ValidateField(t, "udp src", i, 54, api.Uint64ToBytes(uint64(tc["txUdpPortValueList"].([]int32)[reqPacketidx%len(tc["txUdpPortValueList"].([]int32))]), 2))
		cPackets.ValidateField(t, "udp dst", i, 56, api.Uint64ToBytes(uint64(tc["rxUdpPortValueList"].([]int32)[reqPacketidx%len(tc["rxUdpPortValueList"].([]int32))]), 2))
		cPackets.ValidateField(t, "udp length", i, 58, api.Uint64ToBytes(uint64(tc["pktSize"].(int32)-14-4-40), 2))

		expCount := int(tc["pktCount"].(int32))
		actCount := len(cPackets.Packets) - ignoredCount
		if expCount != actCount {
			t.Fatalf("ERROR: expCount %d != actCount %d\n", expCount, actCount)
		}
	}
}
