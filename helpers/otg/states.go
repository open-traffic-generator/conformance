package otg

import (
	"time"

	"github.com/open-traffic-generator/conformance/helpers/table"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func (o *OtgApi) GetIpv4Neighbors() []gosnappi.Neighborsv4State {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting IPv4 Neighbors ...")
	defer o.Timer(time.Now(), "GetIpv4Neighbors")

	sr := api.NewStatesRequest()
	sr.Ipv4Neighbors()
	res, err := api.GetStates(sr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"IPv4 Neighbors",
		[]string{
			"Ethernet Name",
			"IPv4 Address",
			"Link Layer Address",
		},
		15,
	)

	for _, v := range res.Ipv4Neighbors().Items() {
		if v != nil {
			var linkLayerAddress string
			if v.HasLinkLayerAddress() {
				linkLayerAddress = v.LinkLayerAddress()
			} else {
				linkLayerAddress = ""
			}
			tb.AppendRow([]interface{}{
				v.EthernetName(),
				v.Ipv4Address(),
				linkLayerAddress,
			})
		}
	}

	t.Log(tb.String())
	return res.Ipv4Neighbors().Items()
}
