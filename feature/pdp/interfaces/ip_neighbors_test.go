//go:build all || cpdp

package interfaces

import (
	"fmt"
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/dut"
	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestIpNeighbors(t *testing.T) {

	testConst := map[string]interface{}{
		"pktRate":   uint64(50),
		"pktCount":  uint32(100),
		"pktSize":   uint32(128),
		"txMac":     "00:00:01:01:01:01",
		"txIp":      "1.1.1.1",
		"txGateway": "1.1.1.2",
		"txPrefix":  uint32(24),
		"rxMac":     "00:00:01:01:01:02",
		"rxIp":      "2.2.2.1",
		"rxGateway": "2.2.2.2",
		"rxPrefix":  uint32(24),
	}

	api := otg.NewOtgApi(t)

	rmDutConfig := ipNeighborsDutConfig(api, testConst)
	defer rmDutConfig()

	c := ipNeighborsConfig(api, testConst)
	api.SetConfig(c)

	api.WaitFor(
		func() bool { return ipNeighborsIpv4NeighborsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForIpv4Neighbors"},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return ipNeighborsFlowMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func ipNeighborsDutConfig(api *otg.OtgApi, tc map[string]interface{}) func() {
	dc := &api.TestConfig().DutConfigs[0]

	setCfg := fmt.Sprintf(`
		interface %s
			no switchport
			ip address %s/%d
		!
		interface %s
			no switchport
			ip address %s/%d
		!
	`,
		dc.Interfaces[0],
		tc["txGateway"].(string),
		tc["txPrefix"].(uint32),
		dc.Interfaces[1],
		tc["rxGateway"].(string),
		tc["rxPrefix"].(uint32),
	)

	unsetCfg := fmt.Sprintf(`
		interface %s
			no ip address
		!
		interface %s
			no ip address
		!
	`,
		dc.Interfaces[0],
		dc.Interfaces[1],
	)

	return dut.NewDutApi(api.Testing(), dc).SetSshConfig(setCfg, unsetCfg)
}

func ipNeighborsConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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

	flow := c.Flows().Add()
	flow.SetName("ftxV4")
	flow.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
	flow.Rate().SetPps(tc["pktRate"].(uint64))
	flow.Size().SetFixed(tc["pktSize"].(uint32))
	flow.Metrics().SetEnable(true)

	flow.TxRx().Device().
		SetTxNames([]string{dtxIp.Name()}).
		SetRxNames([]string{drxIp.Name()})

	ftxV4Eth := flow.Packet().Add().Ethernet()
	ftxV4Eth.Src().SetValue(dtxEth.Mac())

	ftxV4Ip := flow.Packet().Add().Ipv4()
	ftxV4Ip.Src().SetValue(tc["txIp"].(string))
	ftxV4Ip.Dst().SetValue(tc["rxIp"].(string))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func ipNeighborsFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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

func ipNeighborsIpv4NeighborsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	count := 0
	for _, n := range api.GetIpv4Neighbors() {
		if n.HasLinkLayerAddress() {
			for _, key := range []string{"txGateway", "rxGateway"} {
				if n.Ipv4Address() == tc[key].(string) {
					count += 1
				}
			}
		}
	}

	return count == 2
}
