package udp

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestUdpHeader(t *testing.T) {
	c := otg.Api(t).NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(otg.OtgPorts(t)[0])
	p2 := c.Ports().Add().SetName("p2").SetLocation(otg.OtgPorts(t)[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name(), p2.Name()}).
		SetSpeed(gosnappi.Layer1Speed.SPEED_1_GBPS)

	f1 := c.Flows().Add().SetName("f1")
	f1.TxRx().Port().
		SetTxName(p1.Name()).
		SetRxName(p2.Name())
	f1.Duration().FixedPackets().SetPackets(100)
	f1.Rate().SetPps(50)
	f1.Size().SetFixed(128)
	f1.Metrics().SetEnable(true)

	eth := f1.Packet().Add().Ethernet()
	eth.Src().SetValue("00:00:00:00:00:AA")
	eth.Dst().SetValue("00:00:00:00:00:BB")

	ip := f1.Packet().Add().Ipv4()
	ip.Src().SetValue("1.1.1.10")
	ip.Dst().SetValue("1.1.1.20")

	udp := f1.Packet().Add().Udp()
	udp.SrcPort().SetValue(5000)
	udp.DstPort().SetValue(6000)

	otg.SetConfig(t, c)

	otg.StartTransmit(t)

	err := otg.WaitFor(
		t, udpHeaderMetricsOk, &otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)

	if err != nil {
		t.Fatal(err)
	}
}

func udpHeaderMetricsOk(t *testing.T) (bool, error) {
	m := otg.GetFlowMetrics(t)[0]
	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
			m.FramesTx() == 100 &&
			m.FramesRx() == 100,
		nil
}
