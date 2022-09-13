package otg

import (
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/table"
)

func (o *OtgApi) GetFlowMetrics() []gosnappi.FlowMetric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting flow metrics ...")
	defer o.Timer(time.Now(), "GetFlowMetrics")

	mr := api.NewMetricsRequest()
	mr.Flow()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"Flow Metrics",
		[]string{
			"Name",
			"State",
			"Frames Tx",
			"Frames Rx",
			"FPS Tx",
			"FPS Rx",
			"Bytes Tx",
			"Bytes Rx",
		},
		15,
	)
	for _, v := range res.FlowMetrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.Transmit(),
				v.FramesTx(),
				v.FramesRx(),
				v.FramesTxRate(),
				v.FramesRxRate(),
				v.BytesTx(),
				v.BytesRx(),
			})
		}
	}

	t.Log(tb.String())
	return res.FlowMetrics().Items()
}

func (o *OtgApi) GetBgpv4Metrics() []gosnappi.Bgpv4Metric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting bgpv4 metrics ...")
	defer o.Timer(time.Now(), "GetBgpv4Metrics")

	mr := api.NewMetricsRequest()
	mr.Bgpv4()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"BGPv4 Metrics",
		[]string{
			"Name",
			"State",
			"Routes Adv.",
			"Routes Rec.",
		},
		15,
	)
	for _, v := range res.Bgpv4Metrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.SessionState(),
				v.RoutesAdvertised(),
				v.RoutesReceived(),
			})
		}
	}

	t.Log(tb.String())
	return res.Bgpv4Metrics().Items()
}

func (o *OtgApi) GetLagMetrics() []gosnappi.LagMetric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting lag metrics ...")
	defer o.Timer(time.Now(), "GetLagMetrics")

	mr := api.NewMetricsRequest()
	mr.Lag()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"ISIS Metrics",
		[]string{
			"Name",
			"Oper Status",
		},
		15,
	)
	for _, v := range res.LagMetrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.OperStatus(),
			})
		}
	}

	t.Log(tb.String())
	return res.LagMetrics().Items()
}
