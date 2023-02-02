//go:build all || topo_oc_pdp

package interfaces

import (
	"sync"
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/dut"
	"github.com/open-traffic-generator/conformance/helpers/dut/gnmi"
	"github.com/open-traffic-generator/conformance/helpers/dut/gnmi/gnmipath"
	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestIpv4ForwardingOcPdp(t *testing.T) {

	testConst := map[string]interface{}{
		"pktRate":   int64(50),
		"pktCount":  int32(100),
		"pktSize":   int32(128),
		"txMac":     "00:00:01:01:01:01",
		"txIp":      "1.1.1.1",
		"txGateway": "1.1.1.2",
		"txPrefix":  int32(24),
		"rxMac":     "00:00:01:01:01:02",
		"rxIp":      "2.2.2.1",
		"rxGateway": "2.2.2.2",
		"rxPrefix":  int32(24),
	}

	api := otg.NewOtgApi(t)

	ipv4ForwardingOcPdpDutConfig(api, testConst)

	c := ipv4ForwardingOcPdpConfig(api, testConst)

	api.SetConfig(c)

	api.WaitFor(
		func() bool { return ipv4NeighborsOcPdpOk(api) },
		&otg.WaitForOpts{FnName: "WaitForMacResolution"},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return flowMetricsOcPdpOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func ipv4ForwardingOcPdpDutConfig(api *otg.OtgApi, tc map[string]interface{}) {
	dc := &api.TestConfig().DutConfigs[0]
	t := api.Testing()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		itx := gnmi.Interface{}
		itx.SetName(dc.Interfaces[0])
		itx.SetDescription("DUT ifc connected to OTG tx")
		itx.SetType(gnmi.IETFInterfaces_InterfaceType_ethernetCsmacd)
		itx.SetEnabled(true)

		itxIp := itx.GetOrCreateSubinterface(0).GetOrCreateIpv4()
		itxIp.SetEnabled(true)
		itxIp.GetOrCreateAddress(tc["txGateway"].(string)).
			SetPrefixLength(uint8(tc["txPrefix"].(int32)))

		dutApi := dut.NewDutApi(api.Testing(), dc)
		gnmiClient, err := dut.NewGnmiClient(dutApi)
		if err != nil {
			t.Fatal(err)
		}

		dut.GnmiReplace(gnmiClient, gnmipath.Root().Interface(dc.Interfaces[0]).Config(), &itx)
		itxState := dut.GnmiGet(gnmiClient, gnmipath.Root().Interface(dc.Interfaces[0]).Subinterface(0).Ipv4().Address(tc["txGateway"].(string)).State())
		if *itxState.Ip != tc["txGateway"].(string) || *itxState.PrefixLength != uint8(tc["txPrefix"].(int32)) {
			t.Fatal("itx state did not match expectations")
		}
		wg.Done()
	}()

	go func() {
		irx := gnmi.Interface{}
		irx.SetName(dc.Interfaces[1])
		irx.SetDescription("DUT ifc connected to OTG rx")
		irx.SetType(gnmi.IETFInterfaces_InterfaceType_ethernetCsmacd)
		irx.SetEnabled(true)

		irxIp := irx.GetOrCreateSubinterface(0).GetOrCreateIpv4()
		irxIp.SetEnabled(true)
		irxIp.GetOrCreateAddress(tc["rxGateway"].(string)).
			SetPrefixLength(uint8(tc["rxPrefix"].(int32)))

		dutApi := dut.NewDutApi(api.Testing(), dc)
		gnmiClient, err := dut.NewGnmiClient(dutApi)
		if err != nil {
			t.Fatal(err)
		}

		dut.GnmiReplace(gnmiClient, gnmipath.Root().Interface(dc.Interfaces[1]).Config(), &irx)
		irxState := dut.GnmiGet(gnmiClient, gnmipath.Root().Interface(dc.Interfaces[1]).Subinterface(0).Ipv4().Address(tc["rxGateway"].(string)).State())
		if *irxState.Ip != tc["rxGateway"].(string) || *irxState.PrefixLength != uint8(tc["rxPrefix"].(int32)) {
			t.Fatal("irx state did not match expectations")
		}
		wg.Done()
	}()

	wg.Wait()
}

func ipv4ForwardingOcPdpConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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
	ftxV4Ip.Dst().SetValue(tc["rxIp"].(string))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func flowMetricsOcPdpOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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

func ipv4NeighborsOcPdpOk(api *otg.OtgApi) bool {
	neighbors := api.GetIpv4Neighbors()

	for _, n := range neighbors {
		if !n.HasLinkLayerAddress() {
			return false
		}
	}

	return len(neighbors) > 0

}
