//go:build all || dp3p

package traffic

import (
	"fmt"
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestBasicHexaTraffic(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":  uint64(5),
		"pktCount": uint32(10),
		"pktSize":  uint32(128),
	}

	api := otg.NewOtgApi(t)
	if len(api.TestConfig().OtgPorts) < 6 {
		t.Skipf("Skipping: requires at least 6 OTG ports, got %d", len(api.TestConfig().OtgPorts))
	}

	c := basicHexaTrafficConfig(api, testConst)
	api.SetConfig(c)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return basicHexaTrafficFlowMetricsOk(api, testConst, 3) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	api.StopTransmit()
}

func basicHexaTrafficConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := gosnappi.NewConfig()

	// add 6 ports
	p1 := c.Ports().Add().SetName("p1").SetLocation(api.TestConfig().OtgPorts[0])
	p2 := c.Ports().Add().SetName("p2").SetLocation(api.TestConfig().OtgPorts[1])
	p3 := c.Ports().Add().SetName("p3").SetLocation(api.TestConfig().OtgPorts[2])
	p4 := c.Ports().Add().SetName("p4").SetLocation(api.TestConfig().OtgPorts[3])
	p5 := c.Ports().Add().SetName("p5").SetLocation(api.TestConfig().OtgPorts[4])
	p6 := c.Ports().Add().SetName("p6").SetLocation(api.TestConfig().OtgPorts[5])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name(), p2.Name(), p3.Name(), p4.Name(), p5.Name(), p6.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	// add flows (p1->p4, p2->p5, p3->p6)
	flowPairs := [][2]gosnappi.Port{
		{p1, p4},
		{p2, p5},
		{p3, p6},
	}

	for idx, pair := range flowPairs {
		srcPort := pair[0]
		dstPort := pair[1]

		f := c.Flows().Add().SetName(srcPort.Name() + " -> " + dstPort.Name())
		f.TxRx().Port().SetTxName(srcPort.Name()).SetRxNames([]string{dstPort.Name()})
		f.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
		f.Rate().SetPps(tc["pktRate"].(uint64))
		f.Size().SetFixed(tc["pktSize"].(uint32))
		f.Metrics().SetEnable(true)

		fEth := f.Packet().Add().Ethernet()
		fEth.Src().SetValue(fmt.Sprintf("00:00:0%d:01:01:01", idx+1))
		fEth.Dst().SetValue(fmt.Sprintf("00:00:0%d:02:02:02", idx+1))

		fIp := f.Packet().Add().Ipv4()
		fIp.Src().SetValue(fmt.Sprintf("10.10.%d.1", idx+1))
		fIp.Dst().SetValue(fmt.Sprintf("20.20.%d.1", idx+1))
	}

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func basicHexaTrafficFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}, flowCount int) bool {
	metrics := api.GetFlowMetrics()
	if len(metrics) != flowCount {
		return false
	}

	expCount := uint64(tc["pktCount"].(uint32))
	for _, m := range metrics {
		if m.Transmit() != gosnappi.FlowMetricTransmit.STOPPED {
			return false
		}
		if m.FramesTx() != expCount || m.FramesRx() != expCount {
			return false
		}
	}

	return true
}
