//go:build all || feature || b2b || dp_feature

package tcp

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestTcpHeaderPortIncrDecr(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":        int64(50),
		"pktCount":       int32(100),
		"pktSize":        int32(128),
		"txMac":          "00:00:01:01:01:01",
		"rxMac":          "00:00:01:01:01:02",
		"txIp":           "1.1.1.1",
		"rxIp":           "1.1.1.2",
		"txTcpPortStart": int32(5000),
		"txTcpPortStep":  int32(2),
		"txTcpPortCount": int32(10),
		"rxTcpPortStart": int32(6000),
		"rxTcpPortStep":  int32(2),
		"rxTcpPortCount": int32(10),
	}

	api := otg.NewOtgApi(t)
	c := tcpHeaderPortIncrDecrConfig(api, testConst)

	api.SetConfig(c)

	api.StartCapture()
	api.StartTransmit()

	api.WaitFor(
		func() bool { return tcpHeaderPortIncrDecrMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopCapture()

	tcpHeaderPortIncrDecrCaptureOk(api, c, testConst)
}

func tcpHeaderPortIncrDecrConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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

	ip := f1.Packet().Add().Ipv4()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	tcp := f1.Packet().Add().Tcp()
	tcp.SrcPort().Decrement().
		SetStart(tc["txTcpPortStart"].(int32)).
		SetStep(tc["txTcpPortStep"].(int32)).
		SetCount(tc["txTcpPortCount"].(int32))
	tcp.DstPort().Increment().
		SetStart(tc["rxTcpPortStart"].(int32)).
		SetStep(tc["rxTcpPortStep"].(int32)).
		SetCount(tc["rxTcpPortCount"].(int32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func tcpHeaderPortIncrDecrMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	m := api.GetFlowMetrics()[0]
	expCount := int64(tc["pktCount"].(int32))
	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == expCount &&
		m.FramesRx() == expCount
}

func tcpHeaderPortIncrDecrCaptureOk(api *otg.OtgApi, c gosnappi.Config, tc map[string]interface{}) {
	if !api.TestConfig().OtgCaptureCheck {
		return
	}
	ignoredCount := 0
	txStart := tc["txTcpPortStart"].(int32)
	txStep := tc["txTcpPortStep"].(int32)
	txCount := tc["txTcpPortCount"].(int32)
	rxStart := tc["rxTcpPortStart"].(int32)
	rxStep := tc["rxTcpPortStep"].(int32)
	rxCount := tc["rxTcpPortCount"].(int32)
	cPackets := api.GetCapture(c.Ports().Items()[1].Name())
	t := api.Testing()
	for i := 0; i < len(cPackets.Packets); i++ {
		// ignore unexpected packets based on ethernet src MAC
		if !cPackets.HasField(t, "ethernet src", i, 6, api.MacAddrToBytes(tc["txMac"].(string))) {
			ignoredCount += 1
			continue
		}
		// packet size
		cPackets.ValidateSize(t, i, int(tc["pktSize"].(int32)))
		// ethernet header
		cPackets.ValidateField(t, "ethernet dst", i, 0, api.MacAddrToBytes(tc["rxMac"].(string)))
		cPackets.ValidateField(t, "ethernet type", i, 12, api.Uint64ToBytes(2048, 2))
		// ipv4 header
		cPackets.ValidateField(t, "ipv4 total length", i, 16, api.Uint64ToBytes(uint64(tc["pktSize"].(int32)-14-4), 2))
		cPackets.ValidateField(t, "ipv4 protocol", i, 23, api.Uint64ToBytes(6, 1))
		cPackets.ValidateField(t, "ipv4 src", i, 26, api.Ipv4AddrToBytes(tc["txIp"].(string)))
		cPackets.ValidateField(t, "ipv4 dst", i, 30, api.Ipv4AddrToBytes(tc["rxIp"].(string)))
		// tcp header
		j := int32(i - ignoredCount)
		cPackets.ValidateField(t, "tcp src", i, 34, api.Uint64ToBytes(uint64(txStart-(j%txCount)*txStep), 2))
		cPackets.ValidateField(t, "tcp dst", i, 36, api.Uint64ToBytes(uint64(rxStart+(j%rxCount)*rxStep), 2))
	}
	expCount := int(tc["pktCount"].(int32))
	actCount := len(cPackets.Packets) - ignoredCount
	if expCount != actCount {
		t.Fatalf("ERROR: expCount %d != actCount %d\n", expCount, actCount)
	}
}