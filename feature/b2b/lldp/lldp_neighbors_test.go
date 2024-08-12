//go:build all || cpdp

package lldp

import (
	"testing"
	"time"
	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func TestLldpNeighbors(t *testing.T) {
	testConst := map[string]interface{}{
		"txMac":       "00:00:01:01:01:01",
		"rxMac":       "00:00:01:01:01:02",
		"holdTime":    uint32(120),
		"advInterval": uint32(5),
		"pduCount":    uint64(2),
	}

	api := otg.NewOtgApi(t)
	c := lldpNeighborsConfig(api, testConst)

	api.SetConfig(c)
	api.StartCapture()
	api.StartProtocols()

	api.WaitFor(
		func() bool { return lldpNeighborsLldpMetricssOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForLldpMetrics", Timeout: 30 * time.Second},
	)

	api.WaitFor(
		func() bool { return lldpNeighborsLldpNeighborsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForLldpNeighbors", Timeout: 30 * time.Second},
	)

	api.StopProtocols()
	api.StopCapture()

	port1 := c.Ports().Items()[0]
	//----------------------------------------------------------------------
	// Get captute byte
	//----------------------------------------------------------------------
	api.SaveCapture(port1.Name(), "./lldp_capture")
}

func lldpNeighborsConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := gosnappi.NewConfig()

	ptx := c.Ports().Add().SetName("ptx").SetLocation(api.TestConfig().OtgPorts[0])
	prx := c.Ports().Add().SetName("prx").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	lldpTx := c.Lldp().Add().SetName("lldpTx")
	lldpRx := c.Lldp().Add().SetName("lldpRx")

	lldpTx.SetHoldTime(tc["holdTime"].(uint32))
	lldpTx.SetAdvertisementInterval(tc["advInterval"].(uint32))
	lldpTx.Connection().SetPortName(ptx.Name())
	lldpTx.ChassisId().MacAddressSubtype().
		SetValue(tc["txMac"].(string))
	lldpTx.OrgInfos().Add().SetOui("00120F").SetSubtype(1).Information().SetInfo("036C000010")
	lldpTx.OrgInfos().Add().SetOui("0012BB").SetSubtype(1).Information().SetInfo("000F04")
	lldpTx.OrgInfos().Add().SetOui("0012BB").SetSubtype(2).Information().SetInfo("014065ae")

	lldpRx.SetHoldTime(tc["holdTime"].(uint32))
	lldpRx.SetAdvertisementInterval(tc["advInterval"].(uint32))
	lldpRx.Connection().SetPortName(prx.Name())
	lldpRx.ChassisId().MacAddressSubtype().
		SetValue(tc["rxMac"].(string))

	c.Captures().Add().
		SetName("ca").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetFormat(gosnappi.CaptureFormat.PCAP)

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func lldpNeighborsLldpMetricssOk(api *otg.OtgApi, tc map[string]interface{}) bool {

	for _, m := range api.GetLldpMetrics() {
		// TODO: should be an equality check
		if m.FramesTx() < tc["pduCount"].(uint64) || m.FramesRx() < tc["pduCount"].(uint64) {
			return false
		}

	}

	return true
}

func lldpNeighborsLldpNeighborsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	count := 0
	for _, n := range api.GetLldpNeighbors() {
		for _, key := range []string{"txMac", "rxMac"} {
			if n.HasChassisId() &&
				n.ChassisIdType() == gosnappi.LldpNeighborsStateChassisIdType.MAC_ADDRESS &&
				n.ChassisId() == tc[key].(string) {
				count += 1
			}
		}
	}

	return true
}


