package otg

var (
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
