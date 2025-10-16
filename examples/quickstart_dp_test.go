//go:build all

package examples

import (
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestQuickstartB2BDp(t *testing.T) {
	// Create a new API handle to make API calls against OTG
	api := gosnappi.NewApi()

	// Set the transport protocol to HTTP
	api.NewHttpTransport().SetLocation("https://localhost:8443")

	// Create a new traffic configuration that will be set on OTG
	config := gosnappi.NewConfig()

	// add ports
	p1 := config.Ports().Add().SetName("p1").SetLocation("veth1")
	p2 := config.Ports().Add().SetName("p2").SetLocation("veth2")

	// add flow
	f1 := config.Flows().Add()
	f1.Metrics().SetEnable(true)
	f1.Duration().FixedPackets().SetPackets(50)
	f1.Rate().SetPps(30)

	// add endpoints and packet description flow 1
	f1.SetName(p1.Name() + " -> " + p2.Name()).
		TxRx().Port().SetTxName(p1.Name()).SetRxName(p2.Name())

	f1Eth := f1.Packet().Add().Ethernet()
	f1Eth.Src().SetValue("00:00:01:01:01:01")
	f1Eth.Dst().SetValue("00:00:02:02:02:02")

	f1Ip := f1.Packet().Add().Ipv4()
	f1Ip.Src().SetValue("10.10.10.1")
	f1Ip.Dst().SetValue("20.20.20.1")

	// Optionally, print JSON representation of config
	if j, err := config.Marshal().ToJson(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("\nCONFIGURATION\n%v\n", j)
	}

	// Push configuration
	if _, err := api.SetConfig(config); err != nil {
		t.Fatal(err)
	}

	// Start traffic
	cs := gosnappi.NewControlState()
	cs.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.START)

	if _, err := api.SetControlState(cs); err != nil {
		t.Fatal(err)
	}

	// Fetch metrics for configured flow
	req := gosnappi.NewMetricsRequest()
	req.Flow().SetFlowNames([]string{f1.Name()})
	// Keep polling until either expectation is met or deadline exceeds
	for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(time.Millisecond * 100) {
		metrics, err := api.GetMetrics(req)
		if err != nil || time.Now().After(deadline) {
			t.Fatalf("err = %v || deadline exceeded", err)
		}
		// print YAML representation of flow metrics
		m := metrics.FlowMetrics().Items()[0]
		t.Logf("\nFLOW METRICS\n%v\n", m)
		if m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED && m.FramesRx() == uint64(f1.Duration().FixedPackets().Packets()) {
			break
		}
	}
}
