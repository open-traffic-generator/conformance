//go:build all || cpdp

package isis

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestIsisLspP2pL12(t *testing.T) {

	testConst := map[string]interface{}{
		"pktRate":           uint64(50),
		"pktCount":          uint32(100),
		"pktSize":           uint32(128),
		"txMac":             "00:00:01:01:01:01",
		"txIp":              "1.1.1.1",
		"txGateway":         "1.1.1.2",
		"txPrefix":          uint32(24),
		"txIpv6":            "1100::1",
		"txv6Gateway":       "1100::2",
		"txv6Prefix":        uint32(64),
		"txIsisSystemId":    "640000000001",
		"txIsisAreaAddress": []string{"490001"},
		"rxMac":             "00:00:01:01:01:02",
		"rxIp":              "1.1.1.2",
		"rxGateway":         "1.1.1.1",
		"rxPrefix":          uint32(24),
		"rxIpv6":            "1100::2",
		"rxv6Gateway":       "1100::1",
		"rxv6Prefix":        uint32(64),
		"rxIsisSystemId":    "650000000001",
		"rxIsisAreaAddress": []string{"490001"},
		"txRouteCount":      uint32(1),
		"rxRouteCount":      uint32(1),
		"txAdvRouteV4":      "10.10.10.1",
		"rxAdvRouteV4":      "20.20.20.1",
		"txAdvRouteV6":      "::10:10:10:01",
		"rxAdvRouteV6":      "::20:20:20:01",
	}

	api := otg.NewOtgApi(t)
	c := isisLspP2pL12Config(api, testConst)

	api.SetConfig(c)

	api.StartProtocols()

	api.WaitFor(
		func() bool { return isisLspP2pL12MetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForIsisMetrics",
			Timeout: time.Duration(30) * time.Second},
	)

	api.WaitFor(
		func() bool { return isisLspP2pL12IsisLspsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForIsisLsps",
			Timeout: time.Duration(30) * time.Second},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return isisLspP2pL12FlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func isisLspP2pL12Config(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()

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

	dtxEth.
		Ipv4Addresses().
		Add().
		SetName("dtxIp").
		SetAddress(tc["txIp"].(string)).
		SetGateway(tc["txGateway"].(string)).
		SetPrefix(tc["txPrefix"].(uint32))

	dtxEth.
		Ipv6Addresses().
		Add().
		SetName("dtxIpv6").
		SetAddress(tc["txIpv6"].(string)).
		SetGateway(tc["txv6Gateway"].(string)).
		SetPrefix(tc["txv6Prefix"].(uint32))

	dtxIsis := dtx.Isis().
		SetSystemId(tc["txIsisSystemId"].(string)).
		SetName("dtxIsis")

	dtxIsis.Basic().
		SetIpv4TeRouterId(tc["txIp"].(string)).
		SetHostname(dtxIsis.Name()).
		SetLearnedLspFilter(true)

	dtxIsis.Advanced().
		SetAreaAddresses(tc["txIsisAreaAddress"].([]string)).
		SetLspRefreshRate(900).
		SetEnableAttachedBit(false)

	dtxIsis.Interfaces().
		Add().
		SetEthName(dtxEth.Name()).
		SetName("dtxIsisInt").
		SetNetworkType(gosnappi.IsisInterfaceNetworkType.POINT_TO_POINT).
		SetLevelType(gosnappi.IsisInterfaceLevelType.LEVEL_1_2).
		L2Settings().
		SetDeadInterval(30).
		SetHelloInterval(10).
		SetPriority(0)

	dtxIsisRrV4 := dtxIsis.
		V4Routes().
		Add().SetName("dtxIsisRr4").SetLinkMetric(10)

	dtxIsisRrV4.Addresses().Add().
		SetAddress(tc["txAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	dtxIsisRrV6 := dtxIsis.
		V6Routes().
		Add().SetName("dtxIsisRr6").SetLinkMetric(10)

	dtxIsisRrV6.Addresses().Add().
		SetAddress(tc["txAdvRouteV6"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	// recieve
	drxEth := drx.Ethernets().
		Add().
		SetName("drxEth").
		SetMac(tc["rxMac"].(string)).
		SetMtu(1500)

	drxEth.Connection().SetPortName(prx.Name())

	drxEth.
		Ipv4Addresses().
		Add().
		SetName("drxIp").
		SetAddress(tc["rxIp"].(string)).
		SetGateway(tc["rxGateway"].(string)).
		SetPrefix(tc["rxPrefix"].(uint32))

	drxEth.
		Ipv6Addresses().
		Add().
		SetName("drxIpv6").
		SetAddress(tc["rxIpv6"].(string)).
		SetGateway(tc["rxv6Gateway"].(string)).
		SetPrefix(tc["rxv6Prefix"].(uint32))

	drxIsis := drx.Isis().
		SetSystemId(tc["rxIsisSystemId"].(string)).
		SetName("drxIsis")

	drxIsis.Basic().
		SetIpv4TeRouterId(tc["rxIp"].(string)).
		SetHostname(drxIsis.Name()).
		SetLearnedLspFilter(true)

	drxIsis.Advanced().
		SetAreaAddresses(tc["rxIsisAreaAddress"].([]string)).
		SetLspRefreshRate(900).
		SetEnableAttachedBit(false)

	drxIsis.Interfaces().
		Add().
		SetEthName(drxEth.Name()).
		SetName("drxIsisInt").
		SetNetworkType(gosnappi.IsisInterfaceNetworkType.POINT_TO_POINT).
		SetLevelType(gosnappi.IsisInterfaceLevelType.LEVEL_1_2).
		L2Settings().
		SetDeadInterval(30).
		SetHelloInterval(10).
		SetPriority(0)

	drxIsisRrV4 := drxIsis.
		V4Routes().
		Add().SetName("drxIsisRr4").SetLinkMetric(10)

	drxIsisRrV4.Addresses().Add().
		SetAddress(tc["rxAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

	drxIsisRrV6 := drxIsis.
		V6Routes().
		Add().SetName("drxIsisRr6").SetLinkMetric(10)

	drxIsisRrV6.Addresses().Add().
		SetAddress(tc["rxAdvRouteV6"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

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
		SetTxNames([]string{dtxIsisRrV4.Name()}).
		SetRxNames([]string{drxIsisRrV4.Name()})

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
		SetTxNames([]string{dtxIsisRrV6.Name()}).
		SetRxNames([]string{drxIsisRrV6.Name()})

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
		SetTxNames([]string{drxIsisRrV4.Name()}).
		SetRxNames([]string{dtxIsisRrV4.Name()})

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
		SetTxNames([]string{drxIsisRrV6.Name()}).
		SetRxNames([]string{dtxIsisRrV6.Name()})

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

func isisLspP2pL12IsisLspsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	lspCount := 0
	for _, m := range api.GetIsisLsps() {
		for _, l := range m.Lsps().Items() {
			for _, id := range []string{"txIsisSystemId", "rxIsisSystemId"} {
				if fmt.Sprintf("%s-00-00", tc[id].(string)) == l.LspId() {
					if l.PduType() == gosnappi.IsisLspStatePduType.LEVEL_1 {
						lspCount += 1
					}
					if l.PduType() == gosnappi.IsisLspStatePduType.LEVEL_2 {
						lspCount += 1
					}
				}
			}
		}
	}
	return lspCount == 4
}

func isisLspP2pL12MetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	for _, m := range api.GetIsIsMetrics() {
		if m.L1SessionsUp() < 1 || m.L2SessionsUp() < 1 ||
			m.L1DatabaseSize() < 1 || m.L2DatabaseSize() < 1 {
			return false
		}
	}
	return true
}

func isisLspP2pL12FlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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
