package otg

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/table"
)

type WaitForOpts struct {
	FnName   string
	Interval time.Duration
	Timeout  time.Duration
}

func Timer(t *testing.T, start time.Time, fnName string) {
	elapsed := time.Since(start)
	t.Logf("Elapsed duration for %s: %d ms", fnName, elapsed.Milliseconds())
}

func WaitFor(t *testing.T, fn func(*testing.T) (bool, error), opts *WaitForOpts) error {
	if opts == nil {
		opts = &WaitForOpts{
			FnName: "WaitFor",
		}
	}
	defer Timer(t, time.Now(), opts.FnName)

	if opts.Interval == 0 {
		opts.Interval = 500 * time.Millisecond
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}

	start := time.Now()
	t.Logf("Waiting for %s ...\n", opts.FnName)

	for {
		done, err := fn(t)
		if err != nil {
			return fmt.Errorf("error waiting for %s: %v", opts.FnName, err)
		}
		if done {
			t.Logf("Done waiting for %s\n", opts.FnName)
			return nil
		}

		if time.Since(start) > opts.Timeout {
			return fmt.Errorf("timeout occurred while waiting for %s", opts.FnName)
		}
		time.Sleep(opts.Interval)
	}
}

func CleanupConfig(t *testing.T) {
	t.Log("Cleaning up config ...")
	SetConfig(t, gosnappi.NewConfig())
}

func GetConfig(t *testing.T) gosnappi.Config {
	t.Log("Getting config ...")
	defer Timer(t, time.Now(), "GetConfig")

	res, err := api.GetConfig()
	LogWrnErr(t, nil, err, true)

	return res
}

func SetConfig(t *testing.T, config gosnappi.Config) {
	t.Log("Setting config ...")
	defer Timer(t, time.Now(), "SetConfig")

	res, err := api.SetConfig(config)
	LogWrnErr(t, res, err, true)
}

func StartProtocols(t *testing.T) {
	t.Log("Starting protocol ...")
	defer Timer(t, time.Now(), "StartProtocols")

	ps := api.NewProtocolState().SetState(gosnappi.ProtocolStateState.START)
	res, err := api.SetProtocolState(ps)
	LogWrnErr(t, res, err, true)
}

func StopProtocols(t *testing.T) {
	t.Log("Stopping protocols ...")
	defer Timer(t, time.Now(), "StopProtocols")

	ps := api.NewProtocolState().SetState(gosnappi.ProtocolStateState.STOP)
	res, err := api.SetProtocolState(ps)
	LogWrnErr(t, res, err, true)
}

func StartTransmit(t *testing.T) {
	t.Log("Starting transmit ...")
	defer Timer(t, time.Now(), "StartTransmit")

	ts := api.NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	res, err := api.SetTransmitState(ts)
	LogWrnErr(t, res, err, true)
}

func StopTransmit(t *testing.T) {
	t.Log("Stopping transmit ...")
	defer Timer(t, time.Now(), "StopTransmit")

	ts := api.NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	res, err := api.SetTransmitState(ts)
	LogWrnErr(t, res, err, true)
}

func LogWrnErr(t *testing.T, wrn gosnappi.ResponseWarning, err error, exitOnErr bool) {
	if wrn != nil {
		for _, w := range wrn.Warnings() {
			t.Log("WARNING:", w)
		}
	}

	if err != nil {
		if exitOnErr {
			t.Fatal("ERROR: ", err)
		} else {
			t.Error("ERROR: ", err)
		}
	}
}

func GetPortMetrics(t *testing.T) []gosnappi.PortMetric {
	t.Log("Getting port metrics ...")
	defer Timer(t, time.Now(), "GetPortMetrics")

	mr := api.NewMetricsRequest()
	mr.Port()
	res, err := api.GetMetrics(mr)
	LogWrnErr(t, nil, err, true)

	var out strings.Builder

	border := strings.Repeat("-", 20*4+5)
	out.WriteString("\n")
	out.WriteString(border)
	out.WriteString("\nPort Metrics\n")
	out.WriteString(border)
	out.WriteString("\n")

	for _, rowName := range portMetricRowNames {
		out.WriteString(fmt.Sprintf("%-15s", rowName))
	}
	out.WriteString("\n")

	for _, v := range res.PortMetrics().Items() {
		if v != nil {
			out.WriteString(fmt.Sprintf("%-15v", v.Name()))
			out.WriteString(fmt.Sprintf("%-15v", v.FramesTx()))
			out.WriteString(fmt.Sprintf("%-15v", v.FramesRx()))
			out.WriteString(fmt.Sprintf("%-15v", v.BytesTx()))
			out.WriteString(fmt.Sprintf("%-15v", v.BytesRx()))
		}
		out.WriteString("\n")
	}

	out.WriteString(border)
	out.WriteString("\n\n")
	t.Log(out.String())

	return res.PortMetrics().Items()
}

func GetFlowMetrics(t *testing.T) []gosnappi.FlowMetric {
	t.Log("Getting flow metrics ...")
	defer Timer(t, time.Now(), "GetFlowMetrics")

	mr := api.NewMetricsRequest()
	mr.Flow()
	res, err := api.GetMetrics(mr)
	LogWrnErr(t, nil, err, true)

	tb := table.NewTable(
		"Flow Metrics",
		[]string{
			"Name", "State", "Frames Tx", "Frames Rx", "FPS Tx", "FPS Rx",
			"Bytes Tx", "Bytes Rx",
		},
		15,
	)
	for _, v := range res.FlowMetrics().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.Name(), v.Transmit(), v.FramesTx(), v.FramesRx(),
				v.FramesTxRate(), v.FramesRxRate(), v.BytesTx(), v.BytesRx(),
			})
		}
	}

	t.Log(tb.String())
	return res.FlowMetrics().Items()
}

func GetBgpv4Metrics(t *testing.T) []gosnappi.Bgpv4Metric {
	t.Log("Getting bgpv4 metrics ...")
	defer Timer(t, time.Now(), "GetBgpv4Metrics")

	mr := api.NewMetricsRequest()
	mr.Bgpv4()
	res, err := api.GetMetrics(mr)
	LogWrnErr(t, nil, err, true)

	var out strings.Builder

	border := strings.Repeat("-", 20*4+5)
	out.WriteString("\n")
	out.WriteString(border)
	out.WriteString("\nBGPv4 Metrics\n")
	out.WriteString(border)
	out.WriteString("\n")

	for _, rowName := range bgpMetricRowNames {
		out.WriteString(fmt.Sprintf("%-15s", rowName))
	}
	out.WriteString("\n")

	for _, v := range res.Bgpv4Metrics().Items() {
		if v != nil {
			out.WriteString(fmt.Sprintf("%-15v", v.Name()))
			out.WriteString(fmt.Sprintf("%-15v", v.SessionState()))
			out.WriteString(fmt.Sprintf("%-15v", v.RoutesAdvertised()))
			out.WriteString(fmt.Sprintf("%-15v", v.RoutesReceived()))
		}
		out.WriteString("\n")
	}

	out.WriteString(border)
	out.WriteString("\n\n")
	t.Log(out.String())

	return res.Bgpv4Metrics().Items()
}

func GetBgpv4Prefixes(t *testing.T) []gosnappi.BgpPrefixesState {
	t.Log("Getting bgpv4 metrics ...")
	defer Timer(t, time.Now(), "GetBgpv4Prefixes")

	sr := api.NewStatesRequest()
	sr.BgpPrefixes()
	res, err := api.GetStates(sr)
	LogWrnErr(t, nil, err, true)

	var out strings.Builder

	border := strings.Repeat("-", 20*3+5)
	out.WriteString("\n")
	out.WriteString(border)
	out.WriteString("\nBGPv4 Prefixes\n")
	out.WriteString(border)
	out.WriteString("\n")

	for _, rowName := range bgpPrefixRowNames {
		out.WriteString(fmt.Sprintf("%-15s", rowName))
	}
	out.WriteString("\n")

	for _, v := range res.BgpPrefixes().Items() {
		if v != nil {
			for _, vc := range v.Ipv4UnicastPrefixes().Items() {
				out.WriteString(fmt.Sprintf("%-15v", v.BgpPeerName()))
				out.WriteString(fmt.Sprintf("%-15v", vc.Ipv4Address()))
				out.WriteString(fmt.Sprintf("%-15v", vc.Ipv4NextHop()))
				out.WriteString(fmt.Sprintf("%-15v", vc.PrefixLength()))
				out.WriteString("\n")
			}
		}
	}

	out.WriteString(border)
	out.WriteString("\n\n")
	t.Log(out.String())

	return res.BgpPrefixes().Items()
}
