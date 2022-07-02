package udp

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestUdpHeader(t *testing.T) {
	defer otg.Cleanup(t)

	api := otg.Api()

	c := api.NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(otg.OtgPort1Location())
	p2 := c.Ports().Add().SetName("p2").SetLocation(otg.OtgPort2Location())

	f1 := c.Flows().Add().SetName("f1")
	f1.TxRx().Port().
		SetTxName(p1.Name()).
		SetRxName(p2.Name())
	f1.Duration().FixedPackets().SetPackets(1000)
	f1.Rate().SetPps(500)
	f1.Size().SetFixed(128)
	f1.Metrics().SetEnable(true)

	otg.SetConfig(t, c)

	otg.StartTransmit(t)

	err := otg.WaitFor(
		t,
		func() (bool, error) {
			m := otg.GetFlowMetrics(t)[0]
			return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
					m.FramesTx() == 1000 &&
					m.FramesRx() == 1000,
				nil
		},
		&otg.WaitForOpts{
			FnName: "WaitForFlowMetrics",
		},
	)

	if err != nil {
		t.Fatal(err)
	}
}
