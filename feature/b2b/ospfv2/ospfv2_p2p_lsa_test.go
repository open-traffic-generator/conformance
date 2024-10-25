//go:build all

package ospfv2

import (
	"testing"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestOspfv2P2pLsa(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":      uint64(50),
		"pktCount":     uint32(100),
		"pktSize":      uint32(128),
		"txMac":        "00:00:01:01:01:01",
		"txIp":         "1.1.1.1",
		"txGateway":    "1.1.1.2",
		"txPrefix":     uint32(24),
		"rxMac":        "00:00:01:01:01:02",
		"rxIp":         "1.1.1.2",
		"rxGateway":    "1.1.1.1",
		"rxPrefix":     uint32(24),
		"txRouterName": "dtx_ospfv2",
		"rxRouterName": "drx_ospfv2",
		"txRouteCount": uint32(1),
		"rxRouteCount": uint32(1),
		"txAdvRouteV4": "10.10.10.1",
		"rxAdvRouteV4": "20.20.20.1",
	}

	api := otg.NewOtgApi(t)
	c := ospfv2P2pLsaConfig(api, testConst)

	api.SetConfig(c)

	api.StartProtocols()

	api.WaitFor(
		func() bool { return ospfv2P2pLsaMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForOspfv2Metrics",
			Timeout: time.Duration(30) * time.Second},
	)

	api.WaitFor(
		func() bool { return ospfv2P2pLsasOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForOspfv2Lsas",
			Timeout: time.Duration(30) * time.Second},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return ospfv2P2pLsaFlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func ospfv2P2pLsaConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := gosnappi.NewConfig()

	ptx := c.Ports().Add().SetName("ptx").SetLocation(api.TestConfig().OtgPorts[0])
	prx := c.Ports().Add().SetName("prx").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	dtx := c.Devices().Add().SetName("dtx")
	drx := c.Devices().Add().SetName("drx")

	// transmit
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

	dtxOspfv2 := dtx.Ospfv2().
		SetName(tc["txRouterName"].(string))

	dtxOspfv2.RouterId().SetCustom(dtxIp.Address())
	dtxOspfv2.GracefulRestart().SetHelperMode(true)

	dtxOspfv2.
		SetLsaRetransmitTime(4).
		SetLsaRefreshTime(1800).
		SetInterBurstLsuInterval(33).
		SetStoreLsa(true)

	dtxOspfv2.Capabilities().
		SetNpBit(true).
		SetEBit(true)

	dtxOspfv2Int := dtxOspfv2.
		Interfaces().
		Add().
		SetName("dtxOspfv2Int").
		SetIpv4Name(dtxIp.Name())

	dtxOspfv2Int.NetworkType().PointToPoint()

	dtxOspfv2Int.Advanced().
		SetHelloInterval(9).
		SetDeadInterval(36).
		SetPriority(0).
		SetRoutingMetric(0)

	dtxOspfv2RrV4 := dtxOspfv2.
		V4Routes().
		Add().
		SetName("dtxOspfv2RrV4").
		SetMetric(0)

	dtxOspfv2RrV4.
		Addresses().
		Add().
		SetAddress(tc["txAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	dtxOspfv2RrV4.RouteOrigin().
		InterArea().
		Flags().
		SetAFlag(true).
		SetNFlag(true)

	// recieve
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

	drxOspfv2 := drx.Ospfv2().
		SetName(tc["rxRouterName"].(string))

	drxOspfv2.RouterId().SetCustom(drxIp.Address())
	drxOspfv2.GracefulRestart().SetHelperMode(true)

	drxOspfv2.
		SetLsaRetransmitTime(4).
		SetLsaRefreshTime(1800).
		SetInterBurstLsuInterval(33).
		SetStoreLsa(true)

	drxOspfv2.Capabilities().
		SetNpBit(true).
		SetEBit(true)

	drxOspfv2Int := drxOspfv2.
		Interfaces().
		Add().
		SetName("drxOspfv2Int").
		SetIpv4Name(drxIp.Name())

	drxOspfv2Int.NetworkType().PointToPoint()

	drxOspfv2Int.Advanced().
		SetHelloInterval(9).
		SetDeadInterval(36).
		SetPriority(0).
		SetRoutingMetric(0)

	drxOspfv2RrV4 := drxOspfv2.
		V4Routes().
		Add().
		SetName("drxOspfv2RrV4").
		SetMetric(0)

	drxOspfv2RrV4.
		Addresses().
		Add().
		SetAddress(tc["rxAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

	drxOspfv2RrV4.RouteOrigin().
		InterArea().
		Flags().
		SetAFlag(true).
		SetNFlag(true)

	// traffic
	for i := 1; i <= 2; i++ {
		flow := c.Flows().Add()
		flow.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
		flow.Rate().SetPps(tc["pktRate"].(uint64))
		flow.Size().SetFixed(tc["pktSize"].(uint32))
		flow.Metrics().SetEnable(true)
	}

	ftxV4 := c.Flows().Items()[0]
	ftxV4.SetName("ftxV4")
	ftxV4.TxRx().Device().
		SetTxNames([]string{dtxOspfv2RrV4.Name()}).
		SetRxNames([]string{drxOspfv2RrV4.Name()})

	ftxV4Eth := ftxV4.Packet().Add().Ethernet()
	ftxV4Eth.Src().SetValue(dtxEth.Mac())

	ftxV4Ip := ftxV4.Packet().Add().Ipv4()
	ftxV4Ip.Src().SetValue(tc["txAdvRouteV4"].(string))
	ftxV4Ip.Dst().SetValue(tc["rxAdvRouteV4"].(string))

	ftxV4Tcp := ftxV4.Packet().Add().Tcp()
	ftxV4Tcp.SrcPort().SetValue(5000)
	ftxV4Tcp.DstPort().SetValue(6000)

	frxV4 := c.Flows().Items()[1]
	frxV4.SetName("frxV4")
	frxV4.TxRx().Device().
		SetTxNames([]string{drxOspfv2RrV4.Name()}).
		SetRxNames([]string{dtxOspfv2RrV4.Name()})

	frxV4Eth := frxV4.Packet().Add().Ethernet()
	frxV4Eth.Src().SetValue(drxEth.Mac())

	frxV4Ip := frxV4.Packet().Add().Ipv4()
	frxV4Ip.Src().SetValue(tc["rxAdvRouteV4"].(string))
	frxV4Ip.Dst().SetValue(tc["txAdvRouteV4"].(string))

	frxV4Tcp := frxV4.Packet().Add().Tcp()
	frxV4Tcp.SrcPort().SetValue(6000)
	frxV4Tcp.DstPort().SetValue(5000)

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func ospfv2P2pLsasOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	lsaCount := 0

	var advRouterId string
	var nwSummaryLsaId string
	var routerLsaId string
	var routerLsaLinkId string
	var routerLsaLinkData string

	for _, m := range api.GetOspfv2Lsas() {
		if m.RouterName() == tc["txRouterName"] {
			advRouterId = tc["rxIp"].(string)
			nwSummaryLsaId = tc["rxAdvRouteV4"].(string)
			routerLsaId = tc["rxIp"].(string)
			routerLsaLinkId = tc["txIp"].(string)
			routerLsaLinkData = tc["rxIp"].(string)
		}
		if m.RouterName() == tc["rxRouterName"] {
			advRouterId = tc["txIp"].(string)
			nwSummaryLsaId = tc["txAdvRouteV4"].(string)
			routerLsaId = tc["txIp"].(string)
			routerLsaLinkId = tc["rxIp"].(string)
			routerLsaLinkData = tc["txIp"].(string)
		}

		// validate lsas
		nwSummaryLsas := m.NetworkSummaryLsas().Items()
		if len(nwSummaryLsas) == 1 &&
			nwSummaryLsas[0].Metric() == 0 &&
			nwSummaryLsas[0].Header().AdvertisingRouterId() == advRouterId &&
			nwSummaryLsas[0].Header().LsaId() == nwSummaryLsaId {
			lsaCount += 1
		}

		routerLsas := m.RouterLsas().Items()
		if len(routerLsas) == 1 &&
			routerLsas[0].Header().AdvertisingRouterId() == advRouterId &&
			routerLsas[0].Header().LsaId() == routerLsaId &&
			len(routerLsas[0].Links().Items()) == 2 {
			links := routerLsas[0].Links().Items()
			if links[0].Type() == gosnappi.Ospfv2LinkType.POINT_TO_POINT &&
				links[0].Metric() == 0 &&
				links[0].Id() == routerLsaLinkId &&
				links[0].Data() == routerLsaLinkData &&
				links[1].Type() == gosnappi.Ospfv2LinkType.STUB &&
				links[1].Metric() == 0 {
				lsaCount += 1
			}
		}
	}

	return lsaCount == 4
}

func ospfv2P2pLsaMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	for _, m := range api.GetOspfv2Metrics() {
		if m.FullStateCount() < 1 ||
			m.LsaSent() < 2 ||
			m.LsaReceived() < 2 {
			return false
		}
	}
	return true
}

func ospfv2P2pLsaFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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
