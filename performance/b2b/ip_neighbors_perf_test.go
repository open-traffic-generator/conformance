//go:build all || cpdp

package b2b

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestIpNeighborsPerf(t *testing.T) {
	// TODO: add support for IPv6 as well
	testConst := map[string]interface{}{
		"ifcCounts":           []int{1, 2, 3, 4, 5},
		"ifcCount":            1,
		"startStopCount":      10,
		"startStopIntervalMs": 100,
		"gateways":            []string{},
		"ratePercent":         float32(100),
		"pktSize":             uint32(128),
		"txMac":               "00:00:01:01:01:%02X",
		"txIp":                "1.1.1.%d",
		"txGateway":           "1.1.2.%d",
		"txPrefix":            uint32(24),
		"rxMac":               "00:00:01:01:02:%02X",
		"rxIp":                "1.1.2.%d",
		"rxGateway":           "1.1.1.%d",
		"rxPrefix":            uint32(24),
	}

	distTables := []string{}

	for _, ifcCount := range testConst["ifcCounts"].([]int) {
		if ifcCount > 128 {
			t.Fatal("ERROR: Interface count more than 128 is not supported")
			continue
		}

		testConst["ifcCount"] = ifcCount
		testCase := fmt.Sprintf("IpNeighborsIpHeader2Ports%dIfcFlows", 2*ifcCount)

		api := otg.NewOtgApi(t)
		c := ipNeighborsConfig(api, testConst)

		t.Log("TEST CASE: ", testCase)
		for i := 1; i <= api.TestConfig().OtgIterations; i++ {
			t.Logf("ITERATION: %d\n\n", i)

			api.SetConfig(c)

			api.StartProtocols()

			api.WaitFor(
				func() bool { return ipNeighborsIpv4NeighborsOk(api, testConst) },
				&otg.WaitForOpts{FnName: "WaitForIpv4Neighbors"},
			)

			for j := 1; j <= testConst["startStopCount"].(int); j++ {
				api.StartTransmit()
				api.WaitFor(
					func() bool {
						return flowMetricsTransmitStateOk(api, gosnappi.FlowMetricTransmit.STARTED)
					},
					&otg.WaitForOpts{FnName: "WaitForFlowMetricsStart", Interval: 10 * time.Millisecond},
				)
				api.StopTransmit()
				api.WaitFor(
					func() bool {
						return flowMetricsTransmitStateOk(api, gosnappi.FlowMetricTransmit.STOPPED)
					},
					&otg.WaitForOpts{FnName: "WaitForFlowMetricsStopped", Interval: 10 * time.Millisecond},
				)
			}

			api.Plot().AppendZero()
		}

		api.LogPlot(testCase)

		tb, err := api.Plot().ToTable()
		if err != nil {
			t.Fatal("ERROR:", err)
		}
		distTables = append(distTables, tb)
	}

	for _, d := range distTables {
		t.Log(d)
	}
}

func ipNeighborsConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()

	ptx := c.Ports().Add().SetName("ptx").SetLocation(api.TestConfig().OtgPorts[0])
	prx := c.Ports().Add().SetName("prx").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	tc["gateways"] = []string{}
	for i := 1; i <= tc["ifcCount"].(int); i++ {
		txMac := fmt.Sprintf(tc["txMac"].(string), i)

		txIp := fmt.Sprintf(tc["txIp"].(string), i)
		rxIp := fmt.Sprintf(tc["rxIp"].(string), i)
		tc["gateways"] = append(tc["gateways"].([]string), txIp, rxIp)

		perFlowRate := int(tc["ratePercent"].(float32)) / tc["ifcCount"].(int)

		dtx := c.Devices().Add().SetName(fmt.Sprintf("dtx%d", i))
		drx := c.Devices().Add().SetName(fmt.Sprintf("drx%d", i))

		dtxEth := dtx.Ethernets().
			Add().
			SetName(fmt.Sprintf("dtx%dEth", i)).
			SetMac(txMac).
			SetMtu(1500)

		dtxEth.Connection().SetPortName(ptx.Name())

		dtxIp := dtxEth.
			Ipv4Addresses().
			Add().
			SetName(fmt.Sprintf("dtx%dIp", i)).
			SetAddress(txIp).
			SetGateway(fmt.Sprintf(tc["txGateway"].(string), i)).
			SetPrefix(tc["txPrefix"].(uint32))

		drxEth := drx.Ethernets().
			Add().
			SetName(fmt.Sprintf("drx%dEth", i)).
			SetMac(fmt.Sprintf(tc["rxMac"].(string), i)).
			SetMtu(1500)

		drxEth.Connection().SetPortName(prx.Name())

		drxIp := drxEth.
			Ipv4Addresses().
			Add().
			SetName(fmt.Sprintf("drx%dIp", i)).
			SetAddress(rxIp).
			SetGateway(fmt.Sprintf(tc["rxGateway"].(string), i)).
			SetPrefix(tc["rxPrefix"].(uint32))

		flow := c.Flows().Add()
		flow.SetName(fmt.Sprintf("ftx%dV4", i))
		flow.Duration().Continuous()
		flow.Rate().SetPercentage(float32(perFlowRate))
		flow.Size().SetFixed(tc["pktSize"].(uint32))
		flow.Metrics().SetEnable(true)

		flow.TxRx().Device().
			SetTxNames([]string{dtxIp.Name()}).
			SetRxNames([]string{drxIp.Name()})

		ftxV4Eth := flow.Packet().Add().Ethernet()
		ftxV4Eth.Src().SetValue(txMac)

		ftxV4Ip := flow.Packet().Add().Ipv4()
		ftxV4Ip.Src().SetValue(txIp)
		ftxV4Ip.Dst().SetValue(rxIp)
	}

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func flowMetricsTransmitStateOk(api *otg.OtgApi, ts gosnappi.FlowMetricTransmitEnum) bool {
	for _, m := range api.GetFlowMetrics() {
		if m.Transmit() != ts {
			return false
		}
	}

	return true
}

func ipNeighborsIpv4NeighborsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	count := 0
	for _, n := range api.GetIpv4Neighbors() {
		if n.HasLinkLayerAddress() {
			for _, gateway := range tc["gateways"].([]string) {
				if n.Ipv4Address() == gateway {
					count += 1
					break
				}
			}
		}
	}

	return count == 2*tc["ifcCount"].(int)
}
