//go:build all || cpdp

package b2b

import (
	"fmt"
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestEbgpRoutePrefixPerf(t *testing.T) {

	testConst := map[string]interface{}{
		"pktRate":      uint64(50),
		"pktCount":     uint32(100),
		"pktSize":      uint32(128),
		"txMac":        "00:00:01:01:01:01",
		"txIp":         "1.1.1.1",
		"txGateway":    "1.1.1.2",
		"txPrefix":     uint32(24),
		"txAs":         uint32(1111),
		"rxMac":        "00:00:01:01:01:02",
		"rxIp":         "1.1.1.2",
		"rxGateway":    "1.1.1.1",
		"rxPrefix":     uint32(24),
		"rxAs":         uint32(1112),
		"txRouteCount": uint32(1),
		"rxRouteCount": uint32(1),
		"txNextHopV4":  "1.1.1.3",
		"txNextHopV6":  "::01:01:01:03",
		"rxNextHopV4":  "1.1.1.4",
		"rxNextHopV6":  "::01:01:01:04",
		"txAdvRouteV4": "10.10.10.1",
		"rxAdvRouteV4": "20.20.20.1",
		"txAdvRouteV6": "::10:10:10:01",
		"rxAdvRouteV6": "::20:20:20:01",
	}

	distTables := []string{}
	testCase := fmt.Sprintf("Ebgpv4TcpHeader2Ports2Devices4Flows")

	api := otg.NewOtgApi(t)
	c := ebgpRoutePrefixPerfConfig(api, testConst)

	t.Log("TEST CASE: ", testCase)
	for i := 1; i <= api.TestConfig().OtgIterations; i++ {
		t.Logf("ITERATION: %d\n\n", i)

		api.SetConfig(c)

		api.StartProtocols()

		api.WaitFor(
			func() bool { return ebgpRoutePrefixPerfBgpMetricsOk(api, testConst) },
			&otg.WaitForOpts{FnName: "WaitForBgpv4Metrics"},
		)

		api.StartTransmit()

		api.WaitFor(
			func() bool { return ebgpRoutePrefixPerfFlowMetricsOk(api, testConst) },
			&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
		)

		api.Plot().AppendZero()
	}
	api.LogPlot(testCase)

	tb, err := api.Plot().ToTable()
	if err != nil {
		t.Fatal("ERROR:", err)
	}
	distTables = append(distTables, tb)

	for _, d := range distTables {
		t.Log(d)
	}
}

func ebgpRoutePrefixPerfConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()

	ptx := c.Ports().Add().SetName("ptx").SetLocation(api.TestConfig().OtgPorts[0])
	prx := c.Ports().Add().SetName("prx").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	dtx := c.Devices().Add().SetName("dtx")
	drx := c.Devices().Add().SetName("drx")

	dtxEth := dtx.Ethernets().
		Add().
		SetName("dtxEth").
		SetMac(tc["txMac"].(string)).
		SetMtu(1500)

	dtxEth.Connection().SetPortName(ptx.Name())

	dtxIp := dtxEth.
		Ipv4Addresses().
		Add().
		SetName("dtxIp").
		SetAddress(tc["txIp"].(string)).
		SetGateway(tc["txGateway"].(string)).
		SetPrefix(tc["txPrefix"].(uint32))

	dtxBgp := dtx.Bgp().
		SetRouterId(tc["txIp"].(string))

	dtxBgpv4 := dtxBgp.
		Ipv4Interfaces().Add().
		SetIpv4Name(dtxIp.Name())

	dtxBgpv4Peer := dtxBgpv4.
		Peers().
		Add().
		SetAsNumber(tc["txAs"].(uint32)).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP).
		SetPeerAddress(tc["txGateway"].(string)).
		SetName("dtxBgpv4Peer")

	dtxBgpv4PeerRrV4 := dtxBgpv4Peer.
		V4Routes().
		Add().
		SetNextHopIpv4Address(tc["txNextHopV4"].(string)).
		SetName("dtxBgpv4PeerRrV4").
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	dtxBgpv4PeerRrV4.Addresses().Add().
		SetAddress(tc["txAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	dtxBgpv4PeerRrV4.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	dtxBgpv4PeerRrV4.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	dtxBgpv4PeerRrV4AsPath := dtxBgpv4PeerRrV4.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	dtxBgpv4PeerRrV4AsPath.Segments().Add().
		SetAsNumbers([]uint32{1112, 1113}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	dtxBgpv4PeerRrV6 := dtxBgpv4Peer.
		V6Routes().
		Add().
		SetNextHopIpv6Address(tc["txNextHopV6"].(string)).
		SetName("dtxBgpv4PeerRrV6").
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)

	dtxBgpv4PeerRrV6.Addresses().Add().
		SetAddress(tc["txAdvRouteV6"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	dtxBgpv4PeerRrV6.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	dtxBgpv4PeerRrV6.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	dtxBgpv4PeerRrV6AsPath := dtxBgpv4PeerRrV6.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	dtxBgpv4PeerRrV6AsPath.Segments().Add().
		SetAsNumbers([]uint32{1112, 1113}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	drxEth := drx.Ethernets().
		Add().
		SetName("drxEth").
		SetMac(tc["rxMac"].(string)).
		SetMtu(1500)

	drxEth.Connection().SetPortName(prx.Name())

	drxIp := drxEth.
		Ipv4Addresses().
		Add().
		SetName("drxIp").
		SetAddress(tc["rxIp"].(string)).
		SetGateway(tc["rxGateway"].(string)).
		SetPrefix(tc["rxPrefix"].(uint32))

	drxBgp := drx.Bgp().
		SetRouterId(tc["rxIp"].(string))

	drxBgpv4 := drxBgp.
		Ipv4Interfaces().Add().
		SetIpv4Name(drxIp.Name())

	drxBgpv4Peer := drxBgpv4.
		Peers().
		Add().
		SetAsNumber(tc["rxAs"].(uint32)).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP).
		SetPeerAddress(tc["rxGateway"].(string)).
		SetName("drxBgpv4Peer")

	drxBgpv4PeerRrV4 := drxBgpv4Peer.
		V4Routes().
		Add().
		SetNextHopIpv4Address(tc["rxNextHopV4"].(string)).
		SetName("drxBgpv4PeerRrV4").
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	drxBgpv4PeerRrV4.Addresses().Add().
		SetAddress(tc["rxAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

	drxBgpv4PeerRrV4.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	drxBgpv4PeerRrV4.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	drxBgpv4PeerRrV4AsPath := drxBgpv4PeerRrV4.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	drxBgpv4PeerRrV4AsPath.Segments().Add().
		SetAsNumbers([]uint32{4444}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	drxBgpv4PeerRrV6 := drxBgpv4Peer.
		V6Routes().
		Add().
		SetNextHopIpv6Address(tc["rxNextHopV6"].(string)).
		SetName("drxBgpv4PeerRrV6").
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)

	drxBgpv4PeerRrV6.Addresses().Add().
		SetAddress(tc["rxAdvRouteV6"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

	drxBgpv4PeerRrV6.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	drxBgpv4PeerRrV6.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	drxBgpv4PeerRrV6AsPath := drxBgpv4PeerRrV6.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	drxBgpv4PeerRrV6AsPath.Segments().Add().
		SetAsNumbers([]uint32{4444}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	for i := 1; i <= 4; i++ {
		flow := c.Flows().Add()
		flow.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
		flow.Rate().SetPps(tc["pktRate"].(uint64))
		flow.Size().SetFixed(tc["pktSize"].(uint32))
		flow.Metrics().SetEnable(true)
	}

	ftxV4 := c.Flows().Items()[0]
	ftxV4.SetName("ftxV4")
	ftxV4.TxRx().Device().
		SetTxNames([]string{dtxBgpv4PeerRrV4.Name()}).
		SetRxNames([]string{drxBgpv4PeerRrV4.Name()})

	ftxV4Eth := ftxV4.Packet().Add().Ethernet()
	ftxV4Eth.Src().SetValue(dtxEth.Mac())

	ftxV4Ip := ftxV4.Packet().Add().Ipv4()
	ftxV4Ip.Src().SetValue(tc["txAdvRouteV4"].(string))
	ftxV4Ip.Dst().SetValue(tc["rxAdvRouteV4"].(string))

	ftxV4Tcp := ftxV4.Packet().Add().Tcp()
	ftxV4Tcp.SrcPort().SetValue(5000)
	ftxV4Tcp.DstPort().SetValue(6000)

	ftxV6 := c.Flows().Items()[1]
	ftxV6.SetName("ftxV6")
	ftxV6.TxRx().Device().
		SetTxNames([]string{dtxBgpv4PeerRrV6.Name()}).
		SetRxNames([]string{drxBgpv4PeerRrV6.Name()})

	ftxV6Eth := ftxV6.Packet().Add().Ethernet()
	ftxV6Eth.Src().SetValue(dtxEth.Mac())

	ftxV6Ip := ftxV6.Packet().Add().Ipv6()
	ftxV6Ip.Src().SetValue(tc["txAdvRouteV6"].(string))
	ftxV6Ip.Dst().SetValue(tc["rxAdvRouteV6"].(string))

	ftxV6Tcp := ftxV6.Packet().Add().Tcp()
	ftxV6Tcp.SrcPort().SetValue(5000)
	ftxV6Tcp.DstPort().SetValue(6000)

	frxV4 := c.Flows().Items()[2]
	frxV4.SetName("frxV4")
	frxV4.TxRx().Device().
		SetTxNames([]string{drxBgpv4PeerRrV4.Name()}).
		SetRxNames([]string{dtxBgpv4PeerRrV4.Name()})

	frxV4Eth := frxV4.Packet().Add().Ethernet()
	frxV4Eth.Src().SetValue(drxEth.Mac())

	frxV4Ip := frxV4.Packet().Add().Ipv4()
	frxV4Ip.Src().SetValue(tc["rxAdvRouteV4"].(string))
	frxV4Ip.Dst().SetValue(tc["txAdvRouteV4"].(string))

	frxV4Tcp := frxV4.Packet().Add().Tcp()
	frxV4Tcp.SrcPort().SetValue(6000)
	frxV4Tcp.DstPort().SetValue(5000)

	frxV6 := c.Flows().Items()[3]
	frxV6.SetName("frxV6")
	frxV6.TxRx().Device().
		SetTxNames([]string{drxBgpv4PeerRrV6.Name()}).
		SetRxNames([]string{dtxBgpv4PeerRrV6.Name()})

	frxV6Eth := frxV6.Packet().Add().Ethernet()
	frxV6Eth.Src().SetValue(drxEth.Mac())

	frxV6Ip := frxV6.Packet().Add().Ipv6()
	frxV6Ip.Src().SetValue(tc["rxAdvRouteV6"].(string))
	frxV6Ip.Dst().SetValue(tc["txAdvRouteV6"].(string))

	frxV6Tcp := frxV6.Packet().Add().Tcp()
	frxV6Tcp.SrcPort().SetValue(6000)
	frxV6Tcp.DstPort().SetValue(5000)

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func ebgpRoutePrefixPerfBgpMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	for _, m := range api.GetBgpv4Metrics() {
		if m.SessionState() == gosnappi.Bgpv4MetricSessionState.DOWN ||
			m.RoutesAdvertised() != 2*uint64(tc["txRouteCount"].(uint32)) ||
			m.RoutesReceived() != 2*uint64(tc["rxRouteCount"].(uint32)) {
			return false
		}
	}
	return true
}

func ebgpRoutePrefixPerfFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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
