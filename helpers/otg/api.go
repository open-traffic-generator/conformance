package otg

import (
	"testing"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/plot"
	"github.com/open-traffic-generator/conformance/helpers/testconfig"
	"github.com/open-traffic-generator/snappi/gosnappi"
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
	api.SetVersionCompatibilityCheck(true)
	if tc.OtgGrpcTransport {
		api.NewGrpcTransport().SetLocation(tc.OtgHost).SetRequestTimeout(3600 * time.Second)
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

	cs := o.Api().NewControlState()
	cs.Protocol().All().SetState(gosnappi.StateProtocolAllState.START)
	res, err := o.Api().SetControlState(cs)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StopProtocols() {
	o.Testing().Log("Stopping protocols ...")
	defer o.Timer(time.Now(), "StopProtocols")

	cs := o.Api().NewControlState()
	cs.Protocol().All().SetState(gosnappi.StateProtocolAllState.STOP)
	res, err := o.Api().SetControlState(cs)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StartTransmit() {
	o.Testing().Log("Starting transmit ...")
	defer o.Timer(time.Now(), "StartTransmit")

	cs := o.Api().NewControlState()
	cs.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.START)
	res, err := o.Api().SetControlState(cs)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StopTransmit() {
	o.Testing().Log("Stopping transmit ...")
	defer o.Timer(time.Now(), "StopTransmit")

	cs := o.Api().NewControlState()
	cs.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.STOP)
	res, err := o.Api().SetControlState(cs)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StartCapture() {
	if !o.TestConfig().OtgCaptureCheck {
		o.Testing().Log("Skipped StartCapture")
		return
	}
	o.Testing().Log("Starting capture ...")
	defer o.Timer(time.Now(), "StartCapture")

	cs := o.Api().NewControlState()
	cs.Port().Capture().SetState(gosnappi.StatePortCaptureState.START)
	res, err := o.Api().SetControlState(cs)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) StopCapture() {
	if !o.TestConfig().OtgCaptureCheck {
		o.Testing().Log("Skipped StopCapture")
		return
	}
	o.Testing().Log("Stopping capture ...")
	defer o.Timer(time.Now(), "StopCapture")

	cs := o.Api().NewControlState()
	cs.Port().Capture().SetState(gosnappi.StatePortCaptureState.STOP)
	res, err := o.Api().SetControlState(cs)
	o.LogWrnErr(res, err, true)
}

func (o *OtgApi) NewConfigFromJson(jsonStr string) gosnappi.Config {
	o.Testing().Log("Loading config from JSON ...")
	defer o.Timer(time.Now(), "NewConfigFromJson")

	c := o.Api().NewConfig()
	if err := c.FromJson(jsonStr); err != nil {
		o.Testing().Fatal("ERROR: ", err)
	}

	return c
}

func (o *OtgApi) NewConfigFromYaml(yamlStr string) gosnappi.Config {
	o.Testing().Log("Loading config from YAML ...")
	defer o.Timer(time.Now(), "NewConfigFromYaml")

	c := o.Api().NewConfig()
	if err := c.FromYaml(yamlStr); err != nil {
		o.Testing().Fatal("ERROR: ", err)
	}

	return c
}

func (o *OtgApi) NewConfigFromPbText(pbStr string) gosnappi.Config {
	o.Testing().Log("Loading config from pb text ...")
	defer o.Timer(time.Now(), "NewConfigFromPbText")

	c := o.Api().NewConfig()
	if err := c.FromPbText(pbStr); err != nil {
		o.Testing().Fatal("ERROR: ", err)
	}

	return c
}

func (o *OtgApi) ConfigToJson(config gosnappi.Config) string {
	o.Testing().Log("Serializing config to JSON ...")
	defer o.Timer(time.Now(), "ConfigToJson")

	v, err := config.ToJson()
	if err != nil {
		o.Testing().Fatal("ERROR: ", err)
	}

	return v
}

func (o *OtgApi) ConfigToYaml(config gosnappi.Config) string {
	o.Testing().Log("Serializing config to YAML ...")
	defer o.Timer(time.Now(), "ConfigToYaml")

	v, err := config.ToYaml()
	if err != nil {
		o.Testing().Fatal("ERROR: ", err)
	}

	return v
}

func (o *OtgApi) ConfigToPbText(config gosnappi.Config) string {
	o.Testing().Log("Serializing config to pb text ...")
	defer o.Timer(time.Now(), "ConfigToPbText")

	v, err := config.ToPbText()
	if err != nil {
		o.Testing().Fatal("ERROR: ", err)
	}

	return v
}
