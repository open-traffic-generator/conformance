//go:build all || feature || b2b

package interfaces

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestIpv4Fowarding(t *testing.T) {

	testConst := map[string]interface{}{
		"pktRate":   int64(50),
		"pktCount":  int32(100),
		"pktSize":   int32(128),
		"txMac":     "00:00:01:01:01:01",
		"txIp":      "1.1.1.1",
		"txGateway": "1.1.1.2",
		"txPrefix":  int32(24),
		"rxMac":     "00:00:01:01:01:02",
		"rxIp":      "1.1.1.2",
		"rxGateway": "1.1.1.1",
		"rxPrefix":  int32(24),
	}

	api := otg.NewOtgApi(t)
	c := ipv4ForwardingConfig(api, testConst)

	api.SetConfig(c)

	api.WaitFor(
		func() bool { return macResolutionOk(api) },
		&otg.WaitForOpts{FnName: "WaitForMacResolution"},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return flowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func ipv4ForwardingConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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
		SetPortName(ptx.Name()).
		SetMac(tc["txMac"].(string)).
		SetMtu(1500)

	dtxIp := dtxEth.
		Ipv4Addresses().
		Add().
		SetName("dtxIp").
		SetAddress(tc["txIp"].(string)).
		SetGateway(tc["txGateway"].(string)).
		SetPrefix(tc["txPrefix"].(int32))

	drxEth := drx.Ethernets().
		Add().
		SetName("drxEth").
		SetPortName(prx.Name()).
		SetMac(tc["rxMac"].(string)).
		SetMtu(1500)

	drxIp := drxEth.
		Ipv4Addresses().
		Add().
		SetName("drxIp").
		SetAddress(tc["rxIp"].(string)).
		SetGateway(tc["rxGateway"].(string)).
		SetPrefix(tc["rxPrefix"].(int32))

	flow := c.Flows().Add()
	flow.SetName("ftxV4")
	flow.Duration().FixedPackets().SetPackets(tc["pktCount"].(int32))
	flow.Rate().SetPps(tc["pktRate"].(int64))
	flow.Size().SetFixed(tc["pktSize"].(int32))
	flow.Metrics().SetEnable(true)

	flow.TxRx().Device().
		SetTxNames([]string{dtxIp.Name()}).
		SetRxNames([]string{drxIp.Name()})

	ftxV4Eth := flow.Packet().Add().Ethernet()
	ftxV4Eth.Src().SetValue(dtxEth.Mac())

	ftxV4Ip := flow.Packet().Add().Ipv4()
	ftxV4Ip.Src().SetValue(tc["txIp"].(string))
	ftxV4Ip.Dst().SetValue(tc["txIp"].(string))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func flowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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

func macResolutionOk(api *otg.OtgApi) bool {
	neighbors := api.GetIpv4Neighbors()

	for _, n := range neighbors {
		if !n.HasLinkLayerAddress() {
			return false
		}
	}

	return len(neighbors) > 0

}
