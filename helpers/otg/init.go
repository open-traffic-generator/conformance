package otg

import (
	"github.com/open-traffic-generator/snappi/gosnappi"
)

var (
	api                gosnappi.GosnappiApi
	otgHost            = "https://localhost"
	otgPort1           = "localhost:5555"
	otgPort2           = "localhost:5556"
	portMetricRowNames = []string{
		"Name",
		"Frames Tx",
		"Frames Rx",
		"Bytes Tx",
		"Bytes Rx",
	}
	bgpMetricRowNames = []string{
		"Name",
		"State",
		"Routes Adv.",
		"Routes Rec.",
	}
	bgpPrefixRowNames = []string{
		"Name",
		"IPv4 Addr",
		"IPv4 Next Hop",
		"Prefix Len",
	}
)

type srcDstTestConst struct {
	pktRate      int64
	pktCount     int32
	txRouteCount int32
	rxRouteCount int32
	srcMAC       string
	srcIP        string
	srcGateway   string
	srcPrefix    int32
	srcAS        int32
	dstMAC       string
	dstIP        string
	dstGateway   string
	dstPrefix    int32
	dstAS        int32
}

func Api() gosnappi.GosnappiApi {
	return api
}

func OtgPort1Location() string {
	return otgPort1
}

func OtgPort2Location() string {
	return otgPort2
}

func init() {
	api = gosnappi.NewApi()
	api.NewHttpTransport().SetLocation(otgHost).SetVerify(false)
}
