//go:build all

package examples

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
)

var (
	portA = flag.String("portA", "veth1", "Name of port A interface")
	portZ = flag.String("portZ", "veth2", "Name of port Z interface")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestQuickstartB2BCpDp(t *testing.T) {
	fmt.Println("Running QuickstartB2BCpDp with:")
	fmt.Println("  Ports:", *portA, "<->", *portZ)

	// Create a new API handle to make API calls against OTG
	api := gosnappi.NewApi()

	// Set the transport protocol to HTTP
	api.NewHttpTransport().SetLocation("https://localhost:8443")

	// Create a new traffic configuration that will be set on OTG
	config := gosnappi.NewConfig()

	// add ports
	p1 := config.Ports().Add().SetName("p1").SetLocation(*portA)
	p2 := config.Ports().Add().SetName("p2").SetLocation(*portZ)

	// add devices
	d1 := config.Devices().Add().SetName("d1")
	d2 := config.Devices().Add().SetName("d2")

	// add protocol stacks for device d1
	d1Eth1 := d1.Ethernets().
		Add().
		SetName("d1Eth").
		SetMac("00:00:01:01:01:01").
		SetMtu(1500)
	d1Eth1.Connection().SetPortName(p1.Name())

	d1Eth1.
		Ipv4Addresses().
		Add().
		SetName("p1d1ipv4").
		SetAddress("1.1.1.1").
		SetGateway("1.1.1.2").
		SetPrefix(24)

	d1Bgp := d1.Bgp().
		SetRouterId("1.1.1.1")

	d1BgpIpv4Interface1 := d1Bgp.
		Ipv4Interfaces().Add().
		SetIpv4Name(d1Eth1.Ipv4Addresses().Items()[0].Name())

	d1BgpIpv4Interface1Peer1 := d1BgpIpv4Interface1.
		Peers().
		Add().
		SetAsNumber(1111).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP).
		SetPeerAddress("1.1.1.2").
		SetName("BGP Peer 1")

	d1BgpIpv4Interface1Peer1V4Route1 := d1BgpIpv4Interface1Peer1.
		V4Routes().
		Add().
		SetNextHopIpv4Address("1.1.1.1").
		SetName("p1d1peer1rrv4").
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	d1BgpIpv4Interface1Peer1V4Route1.Addresses().Add().
		SetAddress("10.10.10.0").
		SetPrefix(24).
		SetCount(2).
		SetStep(2)

	d1BgpIpv4Interface1Peer1V4Route1.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	d1BgpIpv4Interface1Peer1V4Route1.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	d1BgpIpv4Interface1Peer1V4Route1AsPath := d1BgpIpv4Interface1Peer1V4Route1.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	d1BgpIpv4Interface1Peer1V4Route1AsPath.Segments().Add().
		SetAsNumbers([]uint32{1112, 1113}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	// add protocol stacks for device d2
	d2Eth1 := d2.Ethernets().
		Add().
		SetName("d2Eth").
		SetMac("00:00:02:02:02:02").
		SetMtu(1500)
	d2Eth1.Connection().SetPortName(p2.Name())

	d2Eth1.
		Ipv4Addresses().
		Add().
		SetName("p2d2ipv4").
		SetAddress("1.1.1.2").
		SetGateway("1.1.1.1").
		SetPrefix(24)

	d2Bgp := d2.Bgp().
		SetRouterId("1.1.1.2")

	d2BgpIpv4Interface1 := d2Bgp.
		Ipv4Interfaces().Add().
		SetIpv4Name(d2Eth1.Ipv4Addresses().Items()[0].Name())

	d2BgpIpv4Interface1Peer1 := d2BgpIpv4Interface1.
		Peers().
		Add().
		SetAsNumber(2222).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP).
		SetPeerAddress("1.1.1.1").
		SetName("BGP Peer 2")

	d2BgpIpv4Interface1Peer1V4Route1 := d2BgpIpv4Interface1Peer1.
		V4Routes().
		Add().
		SetNextHopIpv4Address("1.1.1.2").
		SetName("p2d2peer1rrv4").
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	d2BgpIpv4Interface1Peer1V4Route1.Addresses().Add().
		SetAddress("20.20.20.0").
		SetPrefix(24).
		SetCount(2).
		SetStep(2)

	d2BgpIpv4Interface1Peer1V4Route1.Advanced().
		SetMultiExitDiscriminator(40).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	d2BgpIpv4Interface1Peer1V4Route1.Communities().Add().
		SetAsNumber(100).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	d2BgpIpv4Interface1Peer1V4Route1AsPath := d2BgpIpv4Interface1Peer1V4Route1.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	d2BgpIpv4Interface1Peer1V4Route1AsPath.Segments().Add().
		SetAsNumbers([]uint32{2223, 2224, 2225}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	// add flow
	f1 := config.Flows().Add()
	f1.Metrics().SetEnable(true)
	f1.Duration().FixedPackets().SetPackets(1000)
	f1.Rate().SetPps(500)

	// add endpoints and packet description flow 1
	f1.SetName(p1.Name() + " -> " + p2.Name()).
		TxRx().Device().
		SetTxNames([]string{d1BgpIpv4Interface1Peer1V4Route1.Name()}).
		SetRxNames([]string{d2BgpIpv4Interface1Peer1V4Route1.Name()})

	f1Eth := f1.Packet().Add().Ethernet()
	f1Eth.Src().SetValue(d1Eth1.Mac())
	f1Eth.Dst().Auto()

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

	// Start protocol
	ps := gosnappi.NewControlState()
	ps.Protocol().All().SetState(gosnappi.StateProtocolAllState.START)

	if _, err := api.SetControlState(ps); err != nil {
		t.Fatal(err)
	}

	// Fetch bgpv4 metrics
	req := gosnappi.NewMetricsRequest()
	req.Bgpv4().SetPeerNames([]string{})
	// Keep polling until either expectation is met or deadline exceeds
	for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(time.Millisecond * 100) {
		metrics, err := api.GetMetrics(req)
		if err != nil || time.Now().After(deadline) {
			t.Fatalf("err = %v || deadline exceeded", err)
		}
		// print YAML representation of flow metrics

		allUp := true
		t.Logf("\nBGPv4 METRICS\n")
		for _, m := range metrics.Bgpv4Metrics().Items() {
			t.Logf("%v\n", m)
			if m.SessionState() == gosnappi.Bgpv4MetricSessionState.DOWN || m.RoutesAdvertised() != m.RoutesReceived() {
				allUp = false
				break
			}
		}
		if allUp {
			break
		}
	}

	// Start traffic
	cs := gosnappi.NewControlState()
	cs.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.START)

	if _, err := api.SetControlState(cs); err != nil {
		t.Fatal(err)
	}

	// Fetch metrics for configured flow
	req = gosnappi.NewMetricsRequest()
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
