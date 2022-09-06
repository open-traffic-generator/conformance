package otg

import (
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/plot"
	"github.com/open-traffic-generator/tests/helpers/testconfig"
)

type OtgApi struct {
	t          *testing.T
	testConfig *testconfig.TestConfig
	api        gosnappi.GosnappiApi
	p          *plot.Plot
}

func NewOtgApi(t *testing.T) *OtgApi {
	tc := testconfig.NewTestConfig(t)
	t.Logf("OTG Host: %s\n", tc.OtgHost)
	t.Logf("OTG Port: %v\n", tc.OtgPorts)

	api := gosnappi.NewApi()
	if tc.OtgGrpcTransport {
		api.NewGrpcTransport().SetLocation(tc.OtgHost)
	} else {
		api.NewHttpTransport().SetLocation(tc.OtgHost).SetVerify(false)
	}

	p := plot.NewPlot()

	return &OtgApi{
		t:          t,
		testConfig: tc,
		api:        api,
		p:          p,
	}
}

func (o *OtgApi) TestConfig() *testconfig.TestConfig {
	return o.testConfig
}

func (o *OtgApi) Testing() *testing.T {
	return o.t
}

func (o *OtgApi) Api() gosnappi.GosnappiApi {
	return o.api
}

func (o *OtgApi) Plot() *plot.Plot {
	return o.p
}

func (o *OtgApi) CleanupConfig() {
	o.Testing().Log("Cleaning up config ...")
	o.SetConfig(gosnappi.NewConfig())
}

func (o *OtgApi) GetConfig() gosnappi.Config {
	o.Testing().Log("Getting config ...")
	defer o.Timer(time.Now(), "GetConfig")

	res, err := o.Api().GetConfig()
	o.LogWrnErr(nil, err, true)

	return res
}

func (o *OtgApi) SetConfig(config gosnappi.Config) {
	o.Testing().Log("Setting config ...")
	defer o.Timer(time.Now(), "SetConfig")

	res, err := o.Api().SetConfig(config)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StartProtocols() {
	o.Testing().Log("Starting protocol ...")
	defer o.Timer(time.Now(), "StartProtocols")

	ps := o.Api().NewProtocolState().SetState(gosnappi.ProtocolStateState.START)
	res, err := o.Api().SetProtocolState(ps)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StopProtocols() {
	o.Testing().Log("Stopping protocols ...")
	defer o.Timer(time.Now(), "StopProtocols")

	ps := o.Api().NewProtocolState().SetState(gosnappi.ProtocolStateState.STOP)
	res, err := o.Api().SetProtocolState(ps)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StartTransmit() {
	o.Testing().Log("Starting transmit ...")
	defer o.Timer(time.Now(), "StartTransmit")

	ts := o.Api().NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	res, err := o.Api().SetTransmitState(ts)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StopTransmit() {
	o.Testing().Log("Stopping transmit ...")
	defer o.Timer(time.Now(), "StopTransmit")

	ts := o.Api().NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	res, err := o.Api().SetTransmitState(ts)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StartCapture() {
	if !o.TestConfig().OtgCaptureCheck {
		o.Testing().Log("Skipped StartCapture")
		return
	}
	o.Testing().Log("Starting capture ...")
	defer o.Timer(time.Now(), "StartCapture")

	cs := o.Api().NewCaptureState().SetState(gosnappi.CaptureStateState.START)
	res, err := o.Api().SetCaptureState(cs)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StopCapture() {
	if !o.TestConfig().OtgCaptureCheck {
		o.Testing().Log("Skipped StopCapture")
		return
	}
	o.Testing().Log("Stopping capture ...")
	defer o.Timer(time.Now(), "StopCapture")

	cs := o.Api().NewCaptureState().SetState(gosnappi.CaptureStateState.STOP)
	res, err := o.Api().SetCaptureState(cs)
	o.LogWrnErr(res, err, true)
}
