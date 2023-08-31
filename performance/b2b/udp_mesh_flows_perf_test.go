//go:build all || dp

package b2b

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestUdpMeshFlowsPerf(t *testing.T) {
	testConst := map[string]interface{}{
		"flowCounts": []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048},
		"flowCount":  1,
		"pktRate":    uint64(10),
		"pktCount":   uint32(10),
		"pktSize":    uint32(128),
		"txMac":      "00:00:01:01:01:01",
		"rxMac":      "00:00:01:01:01:02",
		"txIp":       "1.1.1.1",
		"rxIp":       "1.1.1.2",
		"txUdpPort":  uint32(5000),
		"rxUdpPort":  uint32(6000),
	}

	distTables := []string{}

	for _, flowCount := range testConst["flowCounts"].([]int) {
		testConst["flowCount"] = flowCount
		testCase := fmt.Sprintf("UdpHeader2Ports%dFlows", 2*flowCount)

		api := otg.NewOtgApi(t)
		c := udpMeshFlowsPerfConfig(api, testConst)

		t.Log("TEST CASE: ", testCase)
		for i := 1; i <= api.TestConfig().OtgIterations; i++ {
			t.Logf("ITERATION: %d\n\n", i)

			api.SetConfig(c)

			api.StartTransmit()

			api.WaitFor(
				func() bool { return udpHeaderPerfMetricsOk(api, testConst) },
				&otg.WaitForOpts{FnName: "WaitForFlowMetrics", Timeout: 1 * time.Minute},
			)

			api.Plot().AppendZero()
		}

		api.LogPlot(testCase)

		tb, err := api.Plot().ToTable()
		if err != nil {
			t.Fatal("ERROR:", err)
		}
		distTables = append(distTables, tb)
	}

	for _, d := range distTables {
		t.Log(d)
	}
}

func udpMeshFlowsPerfConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(api.TestConfig().OtgPorts[0])
	p2 := c.Ports().Add().SetName("p2").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name(), p2.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	for i := 1; i <= tc["flowCount"].(int); i++ {
		f := c.Flows().Add().SetName(fmt.Sprintf("f%s-%d", p1.Name(), i))
		f.TxRx().Port().
			SetTxName(p1.Name()).
			SetRxName(p2.Name())
		f.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
		f.Rate().SetPps(tc["pktRate"].(uint64))
		f.Size().SetFixed(tc["pktSize"].(uint32))
		f.Metrics().SetEnable(true)

		eth := f.Packet().Add().Ethernet()
		eth.Src().SetValue(tc["txMac"].(string))
		eth.Dst().SetValue(tc["rxMac"].(string))

		ip := f.Packet().Add().Ipv4()
		ip.Src().SetValue(tc["txIp"].(string))
		ip.Dst().SetValue(tc["rxIp"].(string))

		udp := f.Packet().Add().Udp()
		udp.SrcPort().SetValue(tc["txUdpPort"].(uint32))
		udp.DstPort().SetValue(tc["rxUdpPort"].(uint32))
	}

	for i := 1; i <= tc["flowCount"].(int); i++ {
		f := c.Flows().Add().SetName(fmt.Sprintf("f%s-%d", p2.Name(), i))
		f.TxRx().Port().
			SetTxName(p2.Name()).
			SetRxName(p1.Name())
		f.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
		f.Rate().SetPps(tc["pktRate"].(uint64))
		f.Size().SetFixed(tc["pktSize"].(uint32))
		f.Metrics().SetEnable(true)

		eth := f.Packet().Add().Ethernet()
		eth.Src().SetValue(tc["rxMac"].(string))
		eth.Dst().SetValue(tc["txMac"].(string))

		ip := f.Packet().Add().Ipv4()
		ip.Src().SetValue(tc["rxIp"].(string))
		ip.Dst().SetValue(tc["txIp"].(string))

		udp := f.Packet().Add().Udp()
		udp.SrcPort().SetValue(tc["rxUdpPort"].(uint32))
		udp.DstPort().SetValue(tc["txUdpPort"].(uint32))
	}

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func udpHeaderPerfMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	pktCount := uint64(tc["pktCount"].(uint32))
	for _, m := range api.GetFlowMetrics() {
		if m.Transmit() != gosnappi.FlowMetricTransmit.STOPPED ||
			m.FramesTx() != pktCount ||
			m.FramesRx() != pktCount {
			return false
		}
	}

	return true
}
