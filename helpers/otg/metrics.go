package otg

import (
	"time"

	"github.com/open-traffic-generator/conformance/helpers/table"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func (o *OtgApi) GetFlowMetrics() []gosnappi.FlowMetric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting flow metrics ...")
	defer o.Timer(time.Now(), "GetFlowMetrics")

	mr := gosnappi.NewMetricsRequest()
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

	mr := gosnappi.NewMetricsRequest()
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

func (o *OtgApi) GetIsIsMetrics() []gosnappi.IsisMetric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting isis metrics ...")
	defer o.Timer(time.Now(), "GetIsisMetrics")

	mr := gosnappi.NewMetricsRequest()
	mr.Isis()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"ISIS Metrics",
		[]string{
			"Name",
			"L1 Sessions Up",
			"L2 Sessions UP",
			"L1 Database Size",
			"L2 Database Size",
		},
		20,
	)
	for _, v := range res.IsisMetrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.L1SessionsUp(),
				v.L2SessionsUp(),
				v.L1DatabaseSize(),
				v.L2DatabaseSize(),
			})
		}
	}

	t.Log(tb.String())
	return res.IsisMetrics().Items()
}

func (o *OtgApi) GetOspfv2Metrics() []gosnappi.Ospfv2Metric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting ospfv2 metrics ...")
	defer o.Timer(time.Now(), "GetOspfv2Metrics")

	mr := gosnappi.NewMetricsRequest()
	mr.Ospfv2()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"Ospfv2 Metrics",
		[]string{
			"Name",
			"Full State Count",
            "LSA Sent",
            "LSA Received",
		},
		20,
	)
	for _, v := range res.Ospfv2Metrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.FullStateCount(),
				v.LsaSent(),
				v.LsaReceived(),
			})
		}
	}

	t.Log(tb.String())
	return res.Ospfv2Metrics().Items()
}

func (o *OtgApi) GetLldpMetrics() []gosnappi.LldpMetric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting LLDP metrics ...")
	defer o.Timer(time.Now(), "GetLldpMetrics")

	mr := gosnappi.NewMetricsRequest()
	mr.Lldp()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"LLDP Metrics",
		[]string{
			"Name",
			"Frames Tx",
			"Frames Rx",
		},
		15,
	)
	for _, v := range res.LldpMetrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.FramesTx(),
				v.FramesRx(),
			})
		}
	}

	t.Log(tb.String())
	return res.LldpMetrics().Items()
}

func (o *OtgApi) GetLagMetrics() []gosnappi.LagMetric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting LAG metrics ...")
	defer o.Timer(time.Now(), "GetLagMetrics")

	mr := gosnappi.NewMetricsRequest()
	mr.Lag()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"LAG Metrics",
		[]string{
			"Name",
			"Oper Status",
			"Frames Tx",
			"Frames Rx",
			"FPS Tx",
			"FPS Rx",
			"Bytes Tx",
			"Bytes Rx",
		},
		15,
	)
	for _, v := range res.LagMetrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.OperStatus(),
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
	return res.LagMetrics().Items()
}

func (o *OtgApi) GetLacpMetrics() []gosnappi.LacpMetric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting LACP metrics ...")
	defer o.Timer(time.Now(), "GetLacpMetrics")

	mr := gosnappi.NewMetricsRequest()
	mr.Lacp()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"LACP Metrics",
		[]string{
			"LAG Name",
			"LAG Member Port",
			"System ID",
			"Partner ID",
			"LACP Packets Tx",
			"LACP Packets Rx",
		},
		20,
	)
	for _, v := range res.LacpMetrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.LagName(),
				v.LagMemberPortName(),
				v.SystemId(),
				v.PartnerId(),
				v.LacpPacketsTx(),
				v.LacpPacketsRx(),
			})
		}
	}

	t.Log(tb.String())
	return res.LacpMetrics().Items()
}

func (o *OtgApi) GetOspfv3Metrics() []gosnappi.Ospfv3Metric {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting ospfv3 metrics ...")
	defer o.Timer(time.Now(), "GetOspfv3Metrics")

	mr := gosnappi.NewMetricsRequest()
	mr.Ospfv3()
	res, err := api.GetMetrics(mr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"Ospfv3 Metrics",
		[]string{
			"Name",
			"Full State Count",
            "LSA Sent",
            "LSA Received",
		},
		20,
	)
	for _, v := range res.Ospfv3Metrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(),
				v.FullStateCount(),
				v.LsaSent(),
				v.LsaReceived(),
			})
		}
	}

	t.Log(tb.String())
	return res.Ospfv3Metrics().Items()
}
