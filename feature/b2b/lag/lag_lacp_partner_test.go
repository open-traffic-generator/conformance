//go:build all || cpdp

package lldp

import (
	"testing"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestLagLacpPartner(t *testing.T) {
	testConst := map[string]interface{}{
		"txMac":      "00:00:01:01:01:01",
		"rxMac":      "00:00:01:01:01:02",
		"txSystemId": "01:01:01:01:01:01",
		"rxSystemId": "02:02:02:02:02:02",
		"pktRate":    uint64(50),
		"pktCount":   uint32(100),
		"pktSize":    uint32(128),
		"txIp":       "1.1.1.1",
		"rxIp":       "1.1.1.2",
		"txUdpPort":  uint32(5000),
		"rxUdpPort":  uint32(6000),
	}

	api := otg.NewOtgApi(t)
	c := lagLacpPartnerConfig(api, testConst)

	api.SetConfig(c)

	api.StartProtocols()

	api.WaitFor(
		func() bool { return lagLacpPartnerLacpMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForLacpMetrics", Timeout: 30 * time.Second},
	)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return lagLacpPartnerLagMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForLagMetrics", Timeout: 30 * time.Second},
	)
}

func lagLacpPartnerConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()

	ptx := c.Ports().Add().SetName("ptx").SetLocation(api.TestConfig().OtgPorts[0])
	prx := c.Ports().Add().SetName("prx").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	lagTx := c.Lags().Add().SetName("lagTx")
	lagRx := c.Lags().Add().SetName("lagRx")

	lagTx.SetMinLinks(1)
	lagTx.Protocol().Lacp().
		SetActorSystemPriority(1).
		SetActorKey(1).
		SetActorSystemId(tc["txSystemId"].(string))

	lagTxPort := lagTx.Ports().Add().SetPortName(ptx.Name())
	lagTxPort.Lacp().
		SetActorPortNumber(1).
		SetActorPortPriority(1).
		SetActorActivity(gosnappi.LagPortLacpActorActivity.ACTIVE).
		SetLacpduPeriodicTimeInterval(0).
		SetLacpduTimeout(0)

	lagTxPort.Ethernet().
		SetName("lagTxPortEth").
		SetMac(tc["txMac"].(string)).
		SetMtu(1500)

	lagRx.SetMinLinks(1)
	lagRx.Protocol().Lacp().
		SetActorSystemPriority(1).
		SetActorKey(1).
		SetActorSystemId(tc["rxSystemId"].(string))

	lagRxPort := lagRx.Ports().Add().SetPortName(prx.Name())
	lagRxPort.Lacp().
		SetActorPortNumber(1).
		SetActorPortPriority(1).
		SetActorActivity(gosnappi.LagPortLacpActorActivity.ACTIVE).
		SetLacpduPeriodicTimeInterval(0).
		SetLacpduTimeout(0)

	lagRxPort.Ethernet().
		SetName("lagRxPortEth").
		SetMac(tc["rxMac"].(string)).
		SetMtu(1500)

	f1 := c.Flows().Add().SetName("f1")
	f1.TxRx().Port().
		SetTxName(lagTx.Name()).
		SetRxName(lagRx.Name())
	f1.Duration().FixedPackets().SetPackets(tc["pktCount"].(uint32))
	f1.Rate().SetPps(tc["pktRate"].(uint64))
	f1.Size().SetFixed(tc["pktSize"].(uint32))
	f1.Metrics().SetEnable(true)

	eth := f1.Packet().Add().Ethernet()
	eth.Src().SetValue(tc["txMac"].(string))
	eth.Dst().SetValue(tc["rxMac"].(string))

	ip := f1.Packet().Add().Ipv4()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	udp := f1.Packet().Add().Udp()
	udp.SrcPort().SetValue(tc["txUdpPort"].(uint32))
	udp.DstPort().SetValue(tc["rxUdpPort"].(uint32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func lagLacpPartnerLacpMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	count := 0
	for _, m := range api.GetLacpMetrics() {
		if m.SystemId() == tc["txSystemId"].(string) && m.PartnerId() == tc["rxSystemId"].(string) {
			count += 1
		}
		if m.SystemId() == tc["rxSystemId"].(string) && m.PartnerId() == tc["txSystemId"].(string) {
			count += 1
		}
	}

	return count == 2
}

func lagLacpPartnerLagMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	minCount := uint64(tc["pktCount"].(uint32))
	for _, m := range api.GetLagMetrics() {
		if m.OperStatus() != gosnappi.LagMetricOperStatus.UP {
			return false
		}

		if m.Name() == "lagTx" && m.FramesTx() < minCount {
			return false
		}
		if m.Name() == "lagRx" && m.FramesRx() < minCount {
			return false
		}
	}

	return true
}
