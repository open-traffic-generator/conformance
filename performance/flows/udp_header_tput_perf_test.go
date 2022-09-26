//go:build all || perf || b2b || free_perf

package flows

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestUdpHeaderTputPerf(t *testing.T) {
	testConst := map[string]interface{}{
		"pktSizes":  []int32{64, 128, 256, 512, 1024, 1518},
		"lineRates": []float32{50, 100},
		"lineRate":  float32(10),
		"pktCount":  int32(1000000),
		"pktSize":   int32(1518),
		"txMac":     "00:00:01:01:01:01",
		"rxMac":     "00:00:01:01:01:02",
		"txIp":      "1.1.1.1",
		"rxIp":      "1.1.1.2",
		"txUdpPort": int32(5000),
		"rxUdpPort": int32(6000),
	}

	metrics := otg.ThroughputMetrics{
		Metrics: []otg.ThroughputMetric{},
	}

	api := otg.NewOtgApi(t)

	for _, rate := range testConst["lineRates"].([]float32) {
		for _, size := range testConst["pktSizes"].([]int32) {
			testConst["pktSize"] = size
			testConst["lineRate"] = rate
			t.Logf("Test: %d pktSize, %f lineRate\n", size, rate)

			tm := otg.NewThroughputMetric(
				api.Layer1SpeedToMpbs(api.TestConfig().OtgSpeed), rate, uint64(testConst["pktCount"].(int32)), int(size),
			)
			c := udpHeaderTputPerfConfig(api, testConst)

			api.SetConfig(c)

			api.StartTransmit()

			tm.StartCollecting()
			api.WaitFor(
				func() bool { return udpHeaderTputPerfMetricsOk(api, testConst, tm) },
				&otg.WaitForOpts{
					FnName:   "WaitForFlowMetrics",
					Interval: 100 * time.Millisecond,
					Timeout:  10 * time.Minute,
				},
			)
			tm.StopCollecting()

			metrics.Metrics = append(metrics.Metrics, *tm)
		}
	}

	out, err := metrics.ToTable()
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
	t.Log(out)
}

func udpHeaderTputPerfConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(api.TestConfig().OtgPorts[0])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	f := c.Flows().Add().SetName(fmt.Sprintf("f%s", p1.Name()))
	f.TxRx().Port().
		SetTxName(p1.Name())
	f.Duration().FixedPackets().SetPackets(tc["pktCount"].(int32))
	f.Rate().SetPercentage(tc["lineRate"].(float32))
	f.Size().SetFixed(tc["pktSize"].(int32))
	f.Metrics().SetEnable(true)

	eth := f.Packet().Add().Ethernet()
	eth.Src().SetValue(tc["txMac"].(string))
	eth.Dst().SetValue(tc["rxMac"].(string))

	ip := f.Packet().Add().Ipv4()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	udp := f.Packet().Add().Udp()
	udp.SrcPort().SetValue(tc["txUdpPort"].(int32))
	udp.DstPort().SetValue(tc["rxUdpPort"].(int32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func udpHeaderTputPerfMetricsOk(api *otg.OtgApi, tc map[string]interface{}, tm *otg.ThroughputMetric) bool {
	pktCount := int64(tc["pktCount"].(int32))
	for _, m := range api.GetFlowMetrics() {
		tm.AddPpsSnapshot(uint64(m.FramesTx()), int(m.FramesTxRate()))

		if m.Transmit() != gosnappi.FlowMetricTransmit.STOPPED ||
			m.FramesTx() != pktCount {
			return false
		}
	}

	return true
}
