package otg

import (
	"time"

	"github.com/open-traffic-generator/conformance/helpers/plot"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

type WaitForOpts struct {
	FnName   string
	Interval time.Duration
	Timeout  time.Duration
}

func (o *OtgApi) Timer(start time.Time, fnName string) {
	elapsed := time.Since(start)
	o.Plot().AppendDuration(plot.Duration{
		ApiName:  fnName,
		Duration: elapsed,
		Time:     start,
	})
	o.Testing().Logf("Elapsed duration for %s: %d ms", fnName, elapsed.Milliseconds())
}

func (o *OtgApi) WaitFor(fn func() bool, opts *WaitForOpts) {
	t := o.Testing()

	if opts == nil {
		opts = &WaitForOpts{
			FnName: "WaitFor",
		}
	}
	defer o.Timer(time.Now(), opts.FnName)

	if opts.Interval == 0 {
		opts.Interval = 500 * time.Millisecond
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}

	start := time.Now()
	t.Logf("Waiting for %s ...\n", opts.FnName)

	for {

		if fn() {
			t.Logf("Done waiting for %s\n", opts.FnName)
			return
		}

		if time.Since(start) > opts.Timeout {
			t.Fatalf("ERROR: Timeout occurred while waiting for %s\n", opts.FnName)
		}
		time.Sleep(opts.Interval)
	}
}

func (o *OtgApi) LogWrnErr(wrn gosnappi.Warning, err error, exitOnErr bool) {
	t := o.Testing()
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

func (o *OtgApi) LogPlot(name string) {
	t := o.Testing()

	o.Plot().Analyze(name)

	out, err := o.Plot().ToJson()
	if err != nil {
		t.Fatal("ERROR:", err)
	}
	t.Logf("plot: %s\n", out)
}

func (o *OtgApi) Layer1SpeedToMpbs(speed string) int {
	switch gosnappi.Layer1SpeedEnum(speed) {
	case gosnappi.Layer1Speed.SPEED_1_GBPS:
		return 1000
	case gosnappi.Layer1Speed.SPEED_10_GBPS:
		return 10000
	case gosnappi.Layer1Speed.SPEED_25_GBPS:
		return 25000
	case gosnappi.Layer1Speed.SPEED_40_GBPS:
		return 40000
	case gosnappi.Layer1Speed.SPEED_50_GBPS:
		return 50000
	case gosnappi.Layer1Speed.SPEED_100_GBPS:
		return 100000
	case gosnappi.Layer1Speed.SPEED_200_GBPS:
		return 200000
	case gosnappi.Layer1Speed.SPEED_400_GBPS:
		return 400000
	default:
		return 0
	}
}
