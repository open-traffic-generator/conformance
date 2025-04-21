//go:build all

package ospfv3

import (
	"testing"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestOspfv3P2pLsa(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":      uint64(50),
		"pktCount":     uint32(100),
		"pktSize":      uint32(128),
		"txMac":        "00:00:01:01:01:01",
		"txIpV6":       "::1:1:1:1",
		"txGateway":    "::1:1:1:2",
		"txPrefix":     uint32(64),
		"rxMac":        "00:00:01:01:01:02",
		"rxIpV6":       "::1:1:1:2",
		"rxGateway":    "::1:1:1:1",
		"rxPrefix":     uint32(64),
		"txRouterName": "dtx_ospfv3",
		"rxRouterName": "drx_ospfv3",
		"txRouterId":   "5.5.5.5",
		"rxRouterId":   "7.7.7.7",
		"txRouteCount": uint32(1),
		"rxRouteCount": uint32(1),
		"txAdvRouteV6": "4:4:4:0:0:0:0:1",
		"txAddrPrefix": "4:4:4:0:0:0:0:0",
		"rxAdvRouteV6": "6:6:6:0:0:0:0:1",
		"rxAddrPrefix": "6:6:6:0:0:0:0:0",
		"txMetric": 	uint32(10),
		"rxMetric": 	uint32(9),
		"txLinkMetric": uint32(20),
		"rxLinkMetric": uint32(19),
	}

	api := otg.NewOtgApi(t)
	c := ospfv3P2pLsaConfig(api, testConst)

	api.SetConfig(c)

	api.StartProtocols()

	api.WaitFor(
		func() bool { return ospfv3P2pLsaMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForOspfv3Metrics",
			Timeout: time.Duration(30) * time.Second},
	)

	api.WaitFor(
		func() bool { return ospfv3P2pLsasOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForOspfv3Lsas",
			Timeout: time.Duration(30) * time.Second},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return ospfv3P2pLsaFlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

// Please refer to ospfv3 model documentation under 'devices/[ospfv3]' of following url
// for more ospfv3 configuration attributes.
// model: https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/open-traffic-generator/models/master/artifacts/openapi.yaml&nocors#tag/Configuration/operation/set_config

func ospfv3P2pLsaConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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

	dtxIpV6 := dtxEth.
		Ipv6Addresses().
		Add().
		SetName("dtxIpV6").
		SetAddress(tc["txIpV6"].(string)).
		SetGateway(tc["txGateway"].(string)).
		SetPrefix(tc["txPrefix"].(uint32))

	dtxOspfv3 := dtx.Ospfv3()

	dtxOspfv3.RouterId().SetCustom(tc["txRouterId"].(string))

	dtxInstanceOspfv3 := dtxOspfv3.Instances().Add().
		SetName(tc["txRouterName"].(string)).
		SetStoreLsa(true)

	dtxOspfv3Int := dtxInstanceOspfv3.
		Interfaces().
		Add().
		SetName("dtxOspfv3Int").
		SetIpv6Name(dtxIpV6.Name())

	// Note: please change DUT default value for network-type from Broadcast to
	// PointToPoint to make this test interoperable to a port-dut topology
	dtxOspfv3Int.NetworkType().PointToPoint()

	dtxOspfv3Int.Advanced().SetLinkMetric(tc["txLinkMetric"].(uint32))

	dtxOspfv3RrV6 := dtxInstanceOspfv3.
		V6Routes().
		Add().
		SetName("dtxOspfv3RrV6").
		SetMetric(tc["txMetric"].(uint32))

	dtxOspfv3RrV6.
		Addresses().
		Add().
		SetAddress(tc["txAdvRouteV6"].(string)).
		SetPrefix(64).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	// recieve
	drxEth := drx.Ethernets().
		Add().
		SetName("drxEth").
		SetMac(tc["rxMac"].(string)).
		SetMtu(1500)

	drxEth.Connection().SetPortName(prx.Name())

	drxIpV6 := drxEth.
		Ipv6Addresses().
		Add().
		SetName("drxIpV6").
		SetAddress(tc["rxIpV6"].(string)).
		SetGateway(tc["rxGateway"].(string)).
		SetPrefix(tc["rxPrefix"].(uint32))

	drxOspfv3 := drx.Ospfv3()

	drxOspfv3.RouterId().SetCustom(tc["rxRouterId"].(string))

	drxInstanceOspfv3 := drxOspfv3.Instances().Add().
		SetName(tc["rxRouterName"].(string)).
		SetStoreLsa(true)

	drxOspfv3Int := drxInstanceOspfv3.
		Interfaces().
		Add().
		SetName("drxOspfv3Int").
		SetIpv6Name(drxIpV6.Name())

	// Note: please change DUT default value for network-type from Broadcast to
	// PointToPoint to make this test interoperable to a port-dut topology
	drxOspfv3Int.NetworkType().PointToPoint()

	drxOspfv3Int.Advanced().SetLinkMetric(tc["rxLinkMetric"].(uint32))

	drxOspfv3RrV6 := drxInstanceOspfv3.
		V6Routes().
		Add().
		SetName("drxOspfv3RrV6").
		SetMetric(tc["rxMetric"].(uint32))

	drxOspfv3RrV6.
		Addresses().
		Add().
		SetAddress(tc["rxAdvRouteV6"].(string)).
		SetPrefix(64).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

	// traffic
	for i := 1; i <= 2; i++ {
		flow := c.Flows().Add()
		flow.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
		flow.Rate().SetPps(tc["pktRate"].(uint64))
		flow.Size().SetFixed(tc["pktSize"].(uint32))
		flow.Metrics().SetEnable(true)
	}

	ftxV6 := c.Flows().Items()[0]
	ftxV6.SetName("ftxV6")
	ftxV6.TxRx().Device().
		SetTxNames([]string{dtxOspfv3RrV6.Name()}).
		SetRxNames([]string{drxOspfv3RrV6.Name()})

	ftxV6Eth := ftxV6.Packet().Add().Ethernet()
	ftxV6Eth.Src().SetValue(dtxEth.Mac())

	ftxV6Ip := ftxV6.Packet().Add().Ipv6()
	ftxV6Ip.Src().SetValue(tc["txAdvRouteV6"].(string))
	ftxV6Ip.Dst().SetValue(tc["rxAdvRouteV6"].(string))

	ftxV6Tcp := ftxV6.Packet().Add().Tcp()
	ftxV6Tcp.SrcPort().SetValue(5000)
	ftxV6Tcp.DstPort().SetValue(6000)

	frxV6 := c.Flows().Items()[1]
	frxV6.SetName("frxV6")
	frxV6.TxRx().Device().
		SetTxNames([]string{drxOspfv3RrV6.Name()}).
		SetRxNames([]string{dtxOspfv3RrV6.Name()})

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

func ospfv3P2pLsaMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	for _, m := range api.GetOspfv3Metrics() {
		if m.FullStateCount() < 1 ||
			m.LsaSent() < 3 ||
			m.LsaReceived() < 3 {
			return false
		}
	}
	return true
}

func ospfv3P2pLsasOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	lsaCount := 0

	var advRouterId string
	var addrPrefix string
	var metric uint32
	var routerLsaNbrRouterId string
	var routerLsaLinkMetric uint32
	routerLsaLinkType := gosnappi.Ospfv3LinkType.POINT_TO_POINT

	for _, m := range api.GetOspfv3Lsas() {
		if m.RouterName() == tc["txRouterName"] {
			advRouterId = tc["rxRouterId"].(string)
			addrPrefix = tc["rxAddrPrefix"].(string)
			routerLsaNbrRouterId = tc["txRouterId"].(string)
			metric = tc["rxMetric"].(uint32)
			routerLsaLinkMetric = tc["rxLinkMetric"].(uint32)
		}
		if m.RouterName() == tc["rxRouterName"] {
			advRouterId = tc["txRouterId"].(string)
			addrPrefix = tc["txAddrPrefix"].(string)
			routerLsaNbrRouterId = tc["rxRouterId"].(string)
			metric = tc["txMetric"].(uint32)
			routerLsaLinkMetric = tc["txLinkMetric"].(uint32)
		}

		// validate lsas
		interAreaPrefixLsas := m.InterAreaPrefixLsas().Items()
		if len(interAreaPrefixLsas) == 1 &&
			interAreaPrefixLsas[0].AddressPrefix() == addrPrefix &&
			interAreaPrefixLsas[0].Header().AdvertisingRouterId() == advRouterId &&
			interAreaPrefixLsas[0].Metric() == metric {
			lsaCount += 1
		}

		linkLsas := m.LinkLsas().Items()
		if len(linkLsas) == 1 &&
			linkLsas[0].Header().AdvertisingRouterId() == advRouterId {
			lsaCount += 1
		}

		routerLsas := m.RouterLsas().Items()
		if len(routerLsas) == 1 &&
			routerLsas[0].Header().AdvertisingRouterId() == advRouterId &&
			routerLsas[0].NeighborRouterId() == routerLsaNbrRouterId &&
			len(routerLsas[0].Links().Items()) == 1 {
			links := routerLsas[0].Links().Items()
			if links[0].Type() == routerLsaLinkType &&
				links[0].Metric() == routerLsaLinkMetric {
				lsaCount += 1
			}
		}
	}

	return lsaCount == 6
}

func ospfv3P2pLsaFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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
