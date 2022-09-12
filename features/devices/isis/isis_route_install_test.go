//go:build all || feature || b2b

package isis

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestIsIsRouteInstall(t *testing.T) {

	testConst := map[string]interface{}{
		"pktRate":      int64(50),
		"pktCount":     int32(100),
		"pktSize":      int32(128),
		"txMac":        "00:00:01:01:01:01",
		"txIp":         "1.1.1.1",
		"txGateway":    "1.1.1.2",
		"txPrefix":     int32(24),
		"txIpv6":       "1100::1",
		"txv6Gateway":  "1100::2",
		"txv6Prefix":   int32(64),
		"rxMac":        "00:00:01:01:01:02",
		"rxIp":         "1.1.1.2",
		"rxGateway":    "1.1.1.1",
		"rxPrefix":     int32(24),
		"rxIpv6":       "1100::2",
		"rxv6Gateway":  "1100::1",
		"rxv6Prefix":   int32(64),
		"txRouteCount": int32(1),
		"rxRouteCount": int32(1),
		"txAdvRouteV4": "10.10.10.1",
		"rxAdvRouteV4": "20.20.20.1",
		"txAdvRouteV6": "::10:10:10:01",
		"rxAdvRouteV6": "::20:20:20:01",
	}

	api := otg.NewOtgApi(t)
	c := isisRouteInstallConfig(api, testConst)

	fmt.Println(c.ToJson())
	api.SetConfig(c)

	api.StartProtocols()

	api.WaitFor(
		func() bool { return isisRouteInstallMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForIsIsMetrics",
			Timeout: time.Duration(20) * time.Second},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return isisRouteInstallFlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func isisRouteInstallConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()

	ptx := c.Ports().Add().SetName("ptx").SetLocation(api.TestConfig().OtgPorts[0])
	prx := c.Ports().Add().SetName("prx").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	tx := c.Devices().Add().SetName("tx")
	rx := c.Devices().Add().SetName("rx")

	// transmit
	txEth := tx.Ethernets().
		Add().
		SetName("txEth").
		SetPortName(ptx.Name()).
		SetMac(tc["txMac"].(string)).
		SetMtu(1500)

	txEth.
		Ipv4Addresses().
		Add().
		SetName("txIp").
		SetAddress(tc["txIp"].(string)).
		SetGateway(tc["txGateway"].(string)).
		SetPrefix(tc["txPrefix"].(int32))

	txEth.
		Ipv6Addresses().
		Add().
		SetName("txIpv6").
		SetAddress(tc["txIpv6"].(string)).
		SetGateway(tc["txv6Gateway"].(string)).
		SetPrefix(tc["txv6Prefix"].(int32))

	txIsis := tx.Isis().
		SetSystemId("640000000001").
		SetName("tx_isis")

	txIsis.Basic().
		SetIpv4TeRouterId(tc["txIp"].(string)).
		SetHostname("ixia-c-port1")

	txIsis.Advanced().
		SetAreaAddresses([]string{"490001"}).
		SetLspRefreshRate(900).
		SetEnableAttachedBit(false)

	txIsis.Interfaces().
		Add().
		SetEthName(txEth.Name()).
		SetName("tx_isis_int").
		SetNetworkType(gosnappi.IsisInterfaceNetworkType.POINT_TO_POINT).
		SetLevelType(gosnappi.IsisInterfaceLevelType.LEVEL_1_2).
		L2Settings().
		SetDeadInterval(30).
		SetHelloInterval(10).
		SetPriority(0)

	txIsisRrV4 := txIsis.
		V4Routes().
		Add().SetName("tx_isis_rr4").SetLinkMetric(10)

	txIsisRrV4.Addresses().Add().
		SetAddress(tc["txAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(int32)).
		SetStep(1)

	txIsisRrV6 := txIsis.
		V6Routes().
		Add().SetName("tx_isis_rr6").SetLinkMetric(10)

	txIsisRrV6.Addresses().Add().
		SetAddress(tc["txAdvRouteV6"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(int32)).
		SetStep(1)

	// recieve
	rxEth := rx.Ethernets().
		Add().
		SetName("rxEth").
		SetPortName(prx.Name()).
		SetMac(tc["rxMac"].(string)).
		SetMtu(1500)

	rxEth.
		Ipv4Addresses().
		Add().
		SetName("rxIp").
		SetAddress(tc["rxIp"].(string)).
		SetGateway(tc["rxGateway"].(string)).
		SetPrefix(tc["rxPrefix"].(int32))

	rxEth.
		Ipv6Addresses().
		Add().
		SetName("rxIpv6").
		SetAddress(tc["rxIpv6"].(string)).
		SetGateway(tc["rxv6Gateway"].(string)).
		SetPrefix(tc["rxv6Prefix"].(int32))

	rxIsis := rx.Isis().
		SetSystemId("640000000001").
		SetName("rx_isis")

	rxIsis.Basic().
		SetIpv4TeRouterId(tc["rxIp"].(string)).
		SetHostname("ixia-c-port2")

	rxIsis.Advanced().
		SetAreaAddresses([]string{"490001"}).
		SetLspRefreshRate(900).
		SetEnableAttachedBit(false)

	rxIsis.Interfaces().
		Add().
		SetEthName(rxEth.Name()).
		SetName("rx_isis_int").
		SetNetworkType(gosnappi.IsisInterfaceNetworkType.POINT_TO_POINT).
		SetLevelType(gosnappi.IsisInterfaceLevelType.LEVEL_1_2).
		L2Settings().
		SetDeadInterval(30).
		SetHelloInterval(10).
		SetPriority(0)

	rxIsisRrV4 := rxIsis.
		V4Routes().
		Add().SetName("rx_isis_rr4").SetLinkMetric(10)

	rxIsisRrV4.Addresses().Add().
		SetAddress(tc["rxAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(int32)).
		SetStep(1)

	rxIsisRrV6 := rxIsis.
		V6Routes().
		Add().SetName("rx_isis_rr6").SetLinkMetric(10)

	rxIsisRrV6.Addresses().Add().
		SetAddress(tc["rxAdvRouteV6"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(int32)).
		SetStep(1)

	for i := 1; i <= 4; i++ {
		flow := c.Flows().Add()
		flow.Duration().FixedPackets().SetPackets(tc["pktCount"].(int32))
		flow.Rate().SetPps(tc["pktRate"].(int64))
		flow.Size().SetFixed(tc["pktSize"].(int32))
		flow.Metrics().SetEnable(true)
	}

	ftxV4 := c.Flows().Items()[0]
	ftxV4.SetName("ftxV4")
	ftxV4.TxRx().Device().
		SetTxNames([]string{txIsisRrV4.Name()}).
		SetRxNames([]string{rxIsisRrV4.Name()})

	ftxV4Eth := ftxV4.Packet().Add().Ethernet()
	ftxV4Eth.Src().SetValue(txEth.Mac())

	ftxV4Ip := ftxV4.Packet().Add().Ipv4()
	ftxV4Ip.Src().SetValue(tc["txAdvRouteV4"].(string))
	ftxV4Ip.Dst().SetValue(tc["rxAdvRouteV4"].(string))

	ftxV4Tcp := ftxV4.Packet().Add().Tcp()
	ftxV4Tcp.SrcPort().SetValue(5000)
	ftxV4Tcp.DstPort().SetValue(6000)

	ftxV6 := c.Flows().Items()[1]
	ftxV6.SetName("ftxV6")
	ftxV6.TxRx().Device().
		SetTxNames([]string{txIsisRrV6.Name()}).
		SetRxNames([]string{rxIsisRrV6.Name()})

	ftxV6Eth := ftxV6.Packet().Add().Ethernet()
	ftxV6Eth.Src().SetValue(txEth.Mac())

	ftxV6Ip := ftxV6.Packet().Add().Ipv6()
	ftxV6Ip.Src().SetValue(tc["txAdvRouteV6"].(string))
	ftxV6Ip.Dst().SetValue(tc["rxAdvRouteV6"].(string))

	ftxV6Tcp := ftxV6.Packet().Add().Tcp()
	ftxV6Tcp.SrcPort().SetValue(5000)
	ftxV6Tcp.DstPort().SetValue(6000)

	frxV4 := c.Flows().Items()[2]
	frxV4.SetName("frxV4")
	frxV4.TxRx().Device().
		SetTxNames([]string{rxIsisRrV4.Name()}).
		SetRxNames([]string{txIsisRrV4.Name()})

	frxV4Eth := frxV4.Packet().Add().Ethernet()
	frxV4Eth.Src().SetValue(rxEth.Mac())

	frxV4Ip := frxV4.Packet().Add().Ipv4()
	frxV4Ip.Src().SetValue(tc["rxAdvRouteV4"].(string))
	frxV4Ip.Dst().SetValue(tc["txAdvRouteV4"].(string))

	frxV4Tcp := frxV4.Packet().Add().Tcp()
	frxV4Tcp.SrcPort().SetValue(6000)
	frxV4Tcp.DstPort().SetValue(5000)

	frxV6 := c.Flows().Items()[3]
	frxV6.SetName("frxV6")
	frxV6.TxRx().Device().
		SetTxNames([]string{rxIsisRrV6.Name()}).
		SetRxNames([]string{txIsisRrV6.Name()})

	frxV6Eth := frxV6.Packet().Add().Ethernet()
	frxV6Eth.Src().SetValue(rxEth.Mac())

	frxV6Ip := frxV6.Packet().Add().Ipv6()
	frxV6Ip.Src().SetValue(tc["rxAdvRouteV6"].(string))
	frxV6Ip.Dst().SetValue(tc["txAdvRouteV6"].(string))

	frxV6Tcp := frxV6.Packet().Add().Tcp()
	frxV6Tcp.SrcPort().SetValue(6000)
	frxV6Tcp.DstPort().SetValue(5000)

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func isisRouteInstallMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	for _, m := range api.GetIsIsMetrics() {
		if m.L1SessionsUp() != 1 || m.L2SessionsUp() != 1 ||
			m.L1DatabaseSize() != 1 || m.L2DatabaseSize() != 1 {
			return false
		}
	}
	return true
}

func isisRouteInstallFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	pktCount := int64(tc["pktCount"].(int32))

	for _, m := range api.GetFlowMetrics() {
		if m.Transmit() != gosnappi.FlowMetricTransmit.STOPPED ||
			m.FramesTx() != pktCount ||
			m.FramesRx() != pktCount {
			return false
		}

	}

	return true
}
