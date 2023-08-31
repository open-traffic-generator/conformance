//go:build all

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

func TestIpNeighborsOc(t *testing.T) {

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

	rmDutConfig := ipNeighborsOcDutConfig(api, testConst)
	defer rmDutConfig()

	c := ipNeighborsOcConfig(api, testConst)
	api.SetConfig(c)

	api.WaitFor(
		func() bool { return ipNeighborsOcIpv4NeighborsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForIpv4Neighbors"},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return ipNeighborsOcFlowMetricsOcOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func ipNeighborsOcDutConfig(api *otg.OtgApi, tc map[string]interface{}) func() {
	dc := &api.TestConfig().DutConfigs[0]

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		dutApi := dut.NewDutApi(api.Testing(), dc)
		ipNeighborsOcCreateDutInterface(
			api,
			dutApi,
			dc.Interfaces[0],
			"DUT ifc connected to OTG tx",
			tc["txGateway"].(string),
			uint8(tc["txPrefix"].(uint32)),
		)

		wg.Done()
	}()

	go func() {
		dutApi := dut.NewDutApi(api.Testing(), dc)
		ipNeighborsOcCreateDutInterface(
			api,
			dutApi,
			dc.Interfaces[1],
			"DUT ifc connected to OTG rx",
			tc["rxGateway"].(string),
			uint8(tc["rxPrefix"].(uint32)),
		)

		wg.Done()
	}()

	wg.Wait()

	return func() {
		dutApi := dut.NewDutApi(api.Testing(), dc)
		gnmiClient, err := dut.NewGnmiClient(dutApi)
		if err != nil {
			api.Testing().Fatal(err)
		}

		dut.GnmiDelete(gnmiClient, gnmipath.Root().Interface(tc["txGateway"].(string)).Config())
		dut.GnmiDelete(gnmiClient, gnmipath.Root().Interface(tc["rxGateway"].(string)).Config())
	}
}

func ipNeighborsOcCreateDutInterface(api *otg.OtgApi, dutApi *dut.DutApi, name string, description string, ipv4Addr string, ipv4Prefix uint8) {

	ifc := gnmi.Interface{}
	ifc.SetName(name)
	ifc.SetDescription(description)
	ifc.SetType(gnmi.IETFInterfaces_InterfaceType_ethernetCsmacd)
	ifc.SetEnabled(true)

	ifcIp := ifc.GetOrCreateSubinterface(0).GetOrCreateIpv4()
	ifcIp.SetEnabled(true)
	ifcIp.GetOrCreateAddress(ipv4Addr).
		SetPrefixLength(ipv4Prefix)

	gnmiClient, err := dut.NewGnmiClient(dutApi)
	if err != nil {
		api.Testing().Fatal(err)
	}

	dut.GnmiReplace(gnmiClient, gnmipath.Root().Interface(name).Config(), &ifc)

	api.WaitFor(
		func() bool { return ipNeighborsOcDutInterfaceStateOk(api, gnmiClient, name, ipv4Addr, ipv4Prefix) },
		&otg.WaitForOpts{FnName: "WaitDutInterfaceState"},
	)
}

func ipNeighborsOcDutInterfaceStateOk(api *otg.OtgApi, gnmiClient *dut.GnmiClient, name string, ipv4Addr string, ipv4Prefix uint8) bool {
	ifcState := dut.GnmiGet(gnmiClient, gnmipath.Root().Interface(name).Subinterface(0).Ipv4().Address(ipv4Addr).State())
	if ifcState.Ip == nil || ifcState.PrefixLength == nil {
		return false
	}
	if *ifcState.Ip != ipv4Addr {
		api.Testing().Fatalf("IPv4 address did not match; expected: %s, got: %s\n", ipv4Addr, *ifcState.Ip)
	}
	if *ifcState.PrefixLength != ipv4Prefix {
		api.Testing().Fatalf("IPv4 prefix did not match; expected: %d, got: %d\n", ipv4Prefix, *ifcState.PrefixLength)
	}

	return true
}

func ipNeighborsOcConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
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

func ipNeighborsOcFlowMetricsOcOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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

func ipNeighborsOcIpv4NeighborsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
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
