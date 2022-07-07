package otg

import (
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/plot"
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
			t.Fatalf("Timeout occurred while waiting for %s\n", opts.FnName)
		}
		time.Sleep(opts.Interval)
	}
}

func (o *OtgApi) LogWrnErr(wrn gosnappi.ResponseWarning, err error, exitOnErr bool) {
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
