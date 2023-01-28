package otg

import (
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/table"
)

func (o *OtgApi) GetIpv4State() []gosnappi.Neighborsv4State {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting ipv4 state ...")
	defer o.Timer(time.Now(), "GetIpv4State")

	sr := api.NewStatesRequest()
	sr.Ipv4Neighbors()
	res, err := api.GetStates(sr)
	if err != nil {
		o.LogWrnErr(nil, err, true)
		return nil
	}

	tb := table.NewTable(
		"IPv4 State Info",
		[]string{
			"Ethernet Name",
			"IPv4 Addess",
			"Link Layer Address",
		},
		15,
	)

	for _, v := range res.Ipv4Neighbors().Items() {
		if v != nil {
			tb.AppendRow([]interface{}{
				v.EthernetName(),
				v.Ipv4Address(),
				v.LinkLayerAddress(),
			})
		}
	}

	t.Log(tb.String())
	return res.Ipv4Neighbors().Items()
}
