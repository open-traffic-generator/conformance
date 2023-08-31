//go:build all || dp

package vxlan

import (
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestVxlanInnerIpv6(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":        uint64(50),
		"pktCount":       uint32(100),
		"pktSize":        uint32(256),
		"txMac":          "00:00:01:01:01:01",
		"rxMac":          "00:00:01:01:01:02",
		"innerTxMac":     "00:00:01:01:01:03",
		"innerRxMac":     "00:00:01:01:01:04",
		"txIp":           "1.1.1.1",
		"rxIp":           "1.1.1.2",
		"txIpv6":         "::3",
		"rxIpv6":         "::5",
		"txUdpPortValue": uint32(4789),
		"rxUdpPortValue": uint32(4789),
		"vxLanVniValues": []uint32{1000, 1001, 1002, 1003, 1004},
		"txTcpPortValue": uint32(80),
		"rxTcpPortValue": uint32(80),
	}

	api := otg.NewOtgApi(t)
	c := vxlanInnerIpv6Config(api, testConst)

	api.SetConfig(c)

	api.StartCapture()
	api.StartTransmit()

	api.WaitFor(
		func() bool { return vxlanInnerIpv6FlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopCapture()

	vxlanInnerIpv6CaptureOk(api, c, testConst)
}

func vxlanInnerIpv6Config(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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
	udp.SrcPort().SetValue(tc["txUdpPortValue"].(uint32))
	udp.DstPort().SetValue(tc["rxUdpPortValue"].(uint32))

	vxlan := f1.Packet().Add().Vxlan()
	vxlan.Vni().SetValues(tc["vxLanVniValues"].([]uint32))

	eth2 := f1.Packet().Add().Ethernet()
	eth2.Src().SetValue(tc["innerTxMac"].(string))
	eth2.Dst().SetValue(tc["innerRxMac"].(string))

	ip6 := f1.Packet().Add().Ipv6()
	ip6.Src().SetValue(tc["txIpv6"].(string))
	ip6.Dst().SetValue(tc["rxIpv6"].(string))

	tcp := f1.Packet().Add().Tcp()
	tcp.SrcPort().SetValue(tc["txTcpPortValue"].(uint32))
	tcp.DstPort().SetValue(tc["rxTcpPortValue"].(uint32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func vxlanInnerIpv6FlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	m := api.GetFlowMetrics()[0]
	expCount := uint64(tc["pktCount"].(uint32))

	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == expCount &&
		m.FramesRx() == expCount
}

func vxlanInnerIpv6CaptureOk(api *otg.OtgApi, c gosnappi.Config, tc map[string]interface{}) {
	if !api.TestConfig().OtgCaptureCheck {
		return
	}
	ignoredCount := 0
	vxLanVniValues := tc["vxLanVniValues"].([]uint32)
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
		cPackets.ValidateField(t, "udp src", i, 34, api.Uint64ToBytes(uint64(tc["txUdpPortValue"].(uint32)), 2))
		cPackets.ValidateField(t, "udp dst", i, 36, api.Uint64ToBytes(uint64(tc["rxUdpPortValue"].(uint32)), 2))
		cPackets.ValidateField(t, "udp length", i, 38, api.Uint64ToBytes(uint64(tc["pktSize"].(uint32)-14-4-20), 2))
		// vxlan header
		j := i - ignoredCount
		cPackets.ValidateField(t, "vxlan vni", i, 46, api.Uint64ToBytes(uint64(vxLanVniValues[j%len(vxLanVniValues)]), 3))
		// inner ethernet header
		cPackets.ValidateField(t, "ethernet dst", i, 50, api.MacAddrToBytes(tc["innerRxMac"].(string)))
		cPackets.ValidateField(t, "ethernet type", i, 62, api.Uint64ToBytes(34525, 2))
		// inner ipv6 header
		cPackets.ValidateField(t, "ipv6 payload length", i, 68, api.Uint64ToBytes(uint64(tc["pktSize"].(uint32)-14-4-20-8-8-14-4-40), 2))
		cPackets.ValidateField(t, "ipv6 next header", i, 70, api.Uint64ToBytes(6, 1))
		cPackets.ValidateField(t, "ipv6 src", i, 72, api.Ipv6AddrToBytes(tc["txIpv6"].(string)))
		cPackets.ValidateField(t, "ipv6 dst", i, 88, api.Ipv6AddrToBytes(tc["rxIpv6"].(string)))
		// inner tcp header
		cPackets.ValidateField(t, "tcp src", i, 104, api.Uint64ToBytes(uint64(tc["txTcpPortValue"].(uint32)), 2))
		cPackets.ValidateField(t, "tcp dst", i, 106, api.Uint64ToBytes(uint64(tc["rxTcpPortValue"].(uint32)), 2))

	}

	expCount := int(tc["pktCount"].(uint32))
	actCount := len(cPackets.Packets) - ignoredCount
	if expCount != actCount {
		t.Fatalf("ERROR: expCount %d != actCount %d\n", expCount, actCount)
	}
}
