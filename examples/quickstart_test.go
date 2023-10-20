//go:build all

package examples

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestQuickstart(t *testing.T) {
	// Create a new API handle to make API calls against OTG
	api := gosnappi.NewApi()

	// Set the transport protocol to HTTP
	api.NewHttpTransport().SetLocation("https://localhost:8443")

	// Create a new traffic configuration that will be set on OTG
	config := api.NewConfig()

	// Add a test port to the configuration
	ptx := config.Ports().Add().SetName("ptx").SetLocation("veth-a")

	// Configure a flow and set previously created test port as one of endpoints
	flow := config.Flows().Add().SetName("f1")
	flow.TxRx().Port().SetTxName(ptx.Name())
	// and enable tracking flow metrics
	flow.Metrics().SetEnable(true)

	// Configure number of packets to transmit for previously configured flow
	flow.Duration().FixedPackets().SetPackets(100)
	// and fixed byte size of all packets in the flow
	flow.Size().SetFixed(128)

	// Configure protocol headers for all packets in the flow
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	ipv4 := pkt.Add().Ipv4()
	udp := pkt.Add().Udp()
	cus := pkt.Add().Custom()

	eth.Dst().SetValue("00:11:22:33:44:55")
	eth.Src().SetValue("00:11:22:33:44:66")

	ipv4.Src().SetValue("10.1.1.1")
	ipv4.Dst().SetValue("20.1.1.1")

	// Configure repeating patterns for source and destination UDP ports
	udp.SrcPort().SetValues([]uint32{5010, 5015, 5020, 5025, 5030})
	udp.DstPort().Increment().SetStart(6010).SetStep(5).SetCount(5)

	// Configure custom bytes (hex string) in payload
	cus.SetBytes(hex.EncodeToString([]byte("..QUICKSTART SNAPPI..")))

	// Optionally, print JSON representation of config
	if j, err := config.ToJson(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("\nCONFIGURATION\n%v\n", j)
	}

	// Push traffic configuration constructed so far to OTG
	if _, err := api.SetConfig(config); err != nil {
		t.Fatal(err)
	}

	// Start transmitting the packets from configured flow
	cs := api.NewControlState()
	cs.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.START)

	if _, err := api.SetControlState(cs); err != nil {
		t.Fatal(err)
	}

	// Fetch metrics for configured flow
	req := api.NewMetricsRequest()
	req.Flow().SetFlowNames([]string{flow.Name()})
	// Keep polling until either expectation is met or deadline exceeds
	for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(time.Millisecond * 100) {
		metrics, err := api.GetMetrics(req)
		if err != nil || time.Now().After(deadline) {
			t.Fatalf("err = %v || deadline exceeded", err)
		}
		// print YAML representation of flow metrics
		m := metrics.FlowMetrics().Items()[0]
		t.Logf("\nFLOW METRICS\n%v\n", m)
		if m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED {
			break
		}
	}
}
