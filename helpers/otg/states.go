package otg

import (
	"fmt"
	"log"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/table"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

func (o *OtgApi) GetIpv4Neighbors() []gosnappi.Neighborsv4State {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting IPv4 Neighbors ...")
	defer o.Timer(time.Now(), "GetIpv4Neighbors")

	sr := gosnappi.NewStatesRequest()
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
		20,
	)

	for _, v := range res.Ipv4Neighbors().Items() {
		if v != nil {
			linkLayerAddress := ""
			if v.HasLinkLayerAddress() {
				linkLayerAddress = v.LinkLayerAddress()
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

func (o *OtgApi) GetBgpPrefixes() []gosnappi.BgpPrefixesState {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting BGP Prefixes ...")
	defer o.Timer(time.Now(), "GetBgpPrefixes")

	sr := gosnappi.NewStatesRequest()
	sr.BgpPrefixes()
	res, err := api.GetStates(sr)
	log.Println(res)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"BGP Prefixes",
		[]string{
			"Name",
			"IPv4 Address",
			"IPv4 Next Hop",
			"IPv6 Address",
			"IPv6 Next Hop",
			"MED",
			"Local Preference",
		},
		20,
	)

	for _, v := range res.BgpPrefixes().Items() {

		for _, w := range v.Ipv4UnicastPrefixes().Items() {
			row := []interface{}{
				v.BgpPeerName(), fmt.Sprintf("%s/%d", w.Ipv4Address(), w.PrefixLength()), w.Ipv4NextHop(), "",
			}

			if w.HasIpv6NextHop() {
				row = append(row, w.Ipv6NextHop())
			} else {
				row = append(row, "")
			}

			if w.HasMultiExitDiscriminator() {
				row = append(row, w.MultiExitDiscriminator())
			} else {
				row = append(row, "")
			}

			if w.HasLocalPreference() {
				row = append(row, w.LocalPreference())
			} else {
				row = append(row, "")
			}

			tb.AppendRow(row)
		}
		for _, w := range v.Ipv6UnicastPrefixes().Items() {
			row := []interface{}{v.BgpPeerName(), ""}

			if w.HasIpv4NextHop() {
				row = append(row, w.Ipv4NextHop())
			} else {
				row = append(row, "")
			}
			row = append(row, fmt.Sprintf("%s/%d", w.Ipv6Address(), w.PrefixLength()), w.Ipv6NextHop())

			if w.HasMultiExitDiscriminator() {
				row = append(row, w.MultiExitDiscriminator())
			} else {
				row = append(row, "")
			}

			if w.HasLocalPreference() {
				row = append(row, w.LocalPreference())
			} else {
				row = append(row, "")
			}
			tb.AppendRow(row)
		}
	}

	t.Log(tb.String())
	return res.BgpPrefixes().Items()
}

func (o *OtgApi) GetIsisLsps() []gosnappi.IsisLspsState {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting ISIS LSPs ...")
	defer o.Timer(time.Now(), "GetIsisLsps")

	sr := gosnappi.NewStatesRequest()
	sr.IsisLsps()
	res, err := api.GetStates(sr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"ISIS LSPs",
		[]string{
			"Name",
			"LSP ID",
			"PDU Type",
			"IS Type",
		},
		30,
	)

	for _, v := range res.IsisLsps().Items() {
		for _, w := range v.Lsps().Items() {
			tb.AppendRow([]interface{}{
				v.IsisRouterName(),
				w.LspId(),
				w.PduType(),
				w.IsType(),
			})
		}
	}

	t.Log(tb.String())
	return res.IsisLsps().Items()
}

func (o *OtgApi) GetOspfv2Lsas() []gosnappi.Ospfv2LsaState {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting OSPFv2 LSAs ...")
	defer o.Timer(time.Now(), "GetOspfv2Lsas")

	sr := gosnappi.NewStatesRequest()
	sr.Ospfv2Lsas()
	res, err := api.GetStates(sr)
	o.LogWrnErr(nil, err, true)

	fmt.Println(res.Marshal().ToJson())
	return res.Ospfv2Lsas().Items()
}

func (o *OtgApi) GetOspfv3Lsas() []gosnappi.Ospfv3LsaState {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting OSPFv3 LSAs ...")
	defer o.Timer(time.Now(), "GetOspfv3Lsas")

	sr := gosnappi.NewStatesRequest()
	sr.Ospfv3Lsas()
	res, err := api.GetStates(sr)
	o.LogWrnErr(nil, err, true)

	fmt.Println(res.Marshal().ToJson())
	return res.Ospfv3Lsas().Items()
}

func (o *OtgApi) GetLldpNeighbors() []gosnappi.LldpNeighborsState {
	t := o.Testing()
	api := o.Api()

	t.Log("Getting LLDP Neighbors ...")
	defer o.Timer(time.Now(), "GetIpv4Neighbors")

	sr := gosnappi.NewStatesRequest()
	sr.LldpNeighbors()
	res, err := api.GetStates(sr)
	o.LogWrnErr(nil, err, true)

	tb := table.NewTable(
		"LLDP Neighbors",
		[]string{
			"LLDP Name",
			"Chassis ID",
			"Chassis ID Type",
			"System Name",
		},
		20,
	)

	for _, v := range res.LldpNeighbors().Items() {
		row := []interface{}{v.LldpName()}
		if v.HasChassisId() {
			row = append(row, v.ChassisId())
		} else {
			row = append(row, "")
		}
		if v.HasChassisIdType() {
			row = append(row, v.ChassisIdType())
		} else {
			row = append(row, "")
		}
		if v.HasSystemName() {
			row = append(row, v.SystemName())
		} else {
			row = append(row, "")
		}

		tb.AppendRow(row)
	}

	t.Log(tb.String())
	return res.LldpNeighbors().Items()
}
