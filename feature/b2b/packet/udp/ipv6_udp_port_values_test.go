//go:build all || dp

package udp

import (
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestIpv6UdpPortValues(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":         uint64(50),
		"pktCount":        uint32(100),
		"pktSize":         uint32(128),
		"txMac":           "00:00:01:01:01:01",
		"rxMac":           "00:00:01:01:01:02",
		"txIp":            "2000::1",
		"rxIp":            "2000::2",
		"txUdpPortValues": []uint32{5000, 5010, 5020, 5030},
		"rxUdpPortValues": []uint32{6000, 6010, 6020, 6030},
	}

	api := otg.NewOtgApi(t)
	c := ipv6UdpPortValuesConfig(api, testConst)

	api.SetConfig(c)

	api.StartCapture()
	api.StartTransmit()

	api.WaitFor(
		func() bool { return ipv6UdpPortValuesFlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopCapture()

	ipv6UdpPortValuesCaptureOk(api, c, testConst)
}

func ipv6UdpPortValuesConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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

	ip := f1.Packet().Add().Ipv6()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	udp := f1.Packet().Add().Udp()
	udp.SrcPort().SetValues(tc["txUdpPortValues"].([]uint32))
	udp.DstPort().SetValues(tc["rxUdpPortValues"].([]uint32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func ipv6UdpPortValuesFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	m := api.GetFlowMetrics()[0]
	expCount := uint64(tc["pktCount"].(uint32))

	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == expCount &&
		m.FramesRx() == expCount
}

func ipv6UdpPortValuesCaptureOk(api *otg.OtgApi, c gosnappi.Config, tc map[string]interface{}) {
	if !api.TestConfig().OtgCaptureCheck {
		return
	}
	ignoredCount := 0
	txUdpPortValues := tc["txUdpPortValues"].([]uint32)
	rxUdpPortValues := tc["rxUdpPortValues"].([]uint32)
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
		cPackets.ValidateField(t, "ethernet type", i, 12, api.Uint64ToBytes(34525, 2))
		// ipv6 header
		cPackets.ValidateField(t, "ipv6 next header", i, 20, api.Uint64ToBytes(17, 1))
		cPackets.ValidateField(t, "ipv6 src", i, 22, api.Ipv6AddrToBytes(tc["txIp"].(string)))
		cPackets.ValidateField(t, "ipv6 dst", i, 38, api.Ipv6AddrToBytes(tc["rxIp"].(string)))
		// udp header
		j := i - ignoredCount
		cPackets.ValidateField(t, "udp src", i, 54, api.Uint64ToBytes(uint64(txUdpPortValues[j%len(txUdpPortValues)]), 2))
		cPackets.ValidateField(t, "udp dst", i, 56, api.Uint64ToBytes(uint64(rxUdpPortValues[j%len(rxUdpPortValues)]), 2))
		cPackets.ValidateField(t, "udp length", i, 58, api.Uint64ToBytes(uint64(tc["pktSize"].(uint32)-14-4-40), 2))
	}

	expCount := int(tc["pktCount"].(uint32))
	actCount := len(cPackets.Packets) - ignoredCount
	if expCount != actCount {
		t.Fatalf("ERROR: expCount %d != actCount %d\n", expCount, actCount)
	}
}
