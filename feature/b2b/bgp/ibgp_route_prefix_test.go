//go:build all || cpdp

package bgp

import (
	"testing"

	"github.com/open-traffic-generator/conformance/helpers/otg"
	"github.com/open-traffic-generator/snappi/gosnappi"
)

/* Demonstrates advertisement , storage and retrieval of BGP
   routes with attributes such as AS Path, Ext. Community, MED and Local Preference .
   Note that iBGP requires
   a)AS numbers to be identical on both sides,
   b) AS Path should NOT be included in Routes containing own AS Path ( to avoid AS Loop),
   c) Local Preference is supported attribute for iBGP sessions
   Flows are excluded in this example to simplify the test.
   For configuration and access of flow stats refer to ebgp_route_prefix_test.go and flow
   configuration and verification is identical for iBGP and eBGP tests.
*/

func TestIbgpRoutePrefix(t *testing.T) {
	testConst := map[string]interface{}{
		"txMac":        "00:00:01:01:01:01",
		"txIp":         "1.1.1.1",
		"txGateway":    "1.1.1.2",
		"txPrefix":     uint32(24),
		"txAs":         uint32(1111),
		"rxMac":        "00:00:01:01:01:02",
		"rxIp":         "1.1.1.2",
		"rxGateway":    "1.1.1.1",
		"rxPrefix":     uint32(24),
		"rxAs":         uint32(1111),
		"txRouteCount": uint32(1),
		"rxRouteCount": uint32(1),
		"txNextHopV4":  "1.1.1.3",
		"txNextHopV6":  "::1:1:1:3",
		"rxNextHopV4":  "1.1.1.4",
		"rxNextHopV6":  "::1:1:1:4",
		"txAdvRouteV4": "10.10.10.1",
		"rxAdvRouteV4": "20.20.20.1",
		"txAdvRouteV6": "::10:10:10:1",
		"rxAdvRouteV6": "::20:20:20:1",
	}

	api := otg.NewOtgApi(t)
	c := ibgpRoutePrefixConfig(api, testConst)

	api.SetConfig(c)

	api.StartProtocols()

	/* Check if BGP sessions are up and expected routes are Txed and Rxed */
	api.WaitFor(
		func() bool { return ibgpRoutePrefixBgpMetricsOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForBgpv4Metrics"},
	)

	/* Check if each BGP session recieved routes with expected attributes */
	api.WaitFor(
		func() bool { return ibgpRoutePrefixBgpPrefixesOk(api, testConst) },
		&otg.WaitForOpts{FnName: "WaitForBgpRoutePrefixes"},
	)
}

func ibgpRoutePrefixConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()

	ptx := c.Ports().Add().SetName("ptx").SetLocation(api.TestConfig().OtgPorts[0])
	prx := c.Ports().Add().SetName("prx").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{ptx.Name(), prx.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	dtx := c.Devices().Add().SetName("dtx")
	drx := c.Devices().Add().SetName("drx")

	dtxEth := dtx.Ethernets().
		Add().
		SetName("dtxEth").
		SetMac(tc["txMac"].(string)).
		SetMtu(1500)

	dtxEth.Connection().SetPortName(ptx.Name())

	dtxIp := dtxEth.
		Ipv4Addresses().
		Add().
		SetName("dtxIp").
		SetAddress(tc["txIp"].(string)).
		SetGateway(tc["txGateway"].(string)).
		SetPrefix(tc["txPrefix"].(uint32))

	dtxBgp := dtx.Bgp().
		SetRouterId(tc["txIp"].(string))

	dtxBgpv4 := dtxBgp.
		Ipv4Interfaces().Add().
		SetIpv4Name(dtxIp.Name())

	dtxBgpv4Peer := dtxBgpv4.
		Peers().
		Add().
		SetAsNumber(tc["txAs"].(uint32)).
		SetAsType(gosnappi.BgpV4PeerAsType.IBGP).
		SetPeerAddress(tc["txGateway"].(string)).
		SetName("dtxBgpv4Peer")

	dtxBgpv4Peer.LearnedInformationFilter().SetUnicastIpv4Prefix(true).SetUnicastIpv6Prefix(true)

	dtxBgpv4PeerRrV4 := dtxBgpv4Peer.
		V4Routes().
		Add().
		SetName("dtxBgpv4PeerRrV4").
		SetNextHopIpv4Address(tc["txNextHopV4"].(string)).
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	dtxBgpv4PeerRrV4.Addresses().Add().
		SetAddress(tc["txAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	dtxBgpv4PeerRrV4.Advanced().
		SetMultiExitDiscriminator(50).
		SetLocalPreference(100).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	dtxBgpv4PeerRrV4.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	dtxBgpv4PeerRrV6 := dtxBgpv4Peer.
		V6Routes().
		Add().
		SetName("dtxBgpv4PeerRrV6").
		SetNextHopIpv6Address(tc["txNextHopV6"].(string)).
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)

	dtxBgpv4PeerRrV6.Addresses().Add().
		SetAddress(tc["txAdvRouteV6"].(string)).
		SetPrefix(128).
		SetCount(tc["txRouteCount"].(uint32)).
		SetStep(1)

	dtxBgpv4PeerRrV6.Advanced().
		SetMultiExitDiscriminator(50).
		SetLocalPreference(100).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	dtxBgpv4PeerRrV6.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	drxEth := drx.Ethernets().
		Add().
		SetName("drxEth").
		SetMac(tc["rxMac"].(string)).
		SetMtu(1500)

	drxEth.Connection().SetPortName(prx.Name())

	drxIp := drxEth.
		Ipv4Addresses().
		Add().
		SetName("drxIp").
		SetAddress(tc["rxIp"].(string)).
		SetGateway(tc["rxGateway"].(string)).
		SetPrefix(tc["rxPrefix"].(uint32))

	drxBgp := drx.Bgp().
		SetRouterId(tc["rxIp"].(string))

	drxBgpv4 := drxBgp.
		Ipv4Interfaces().Add().
		SetIpv4Name(drxIp.Name())

	drxBgpv4Peer := drxBgpv4.
		Peers().
		Add().
		SetAsNumber(tc["rxAs"].(uint32)).
		SetAsType(gosnappi.BgpV4PeerAsType.IBGP).
		SetPeerAddress(tc["rxGateway"].(string)).
		SetName("drxBgpv4Peer")

	drxBgpv4Peer.LearnedInformationFilter().SetUnicastIpv4Prefix(true).SetUnicastIpv6Prefix(true)

	drxBgpv4PeerRrV4 := drxBgpv4Peer.
		V4Routes().
		Add().
		SetName("drxBgpv4PeerRrV4").
		SetNextHopIpv4Address(tc["rxNextHopV4"].(string)).
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	drxBgpv4PeerRrV4.Addresses().Add().
		SetAddress(tc["rxAdvRouteV4"].(string)).
		SetPrefix(32).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

	drxBgpv4PeerRrV4.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	drxBgpv4PeerRrV4.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	drxBgpv4PeerRrV6 := drxBgpv4Peer.
		V6Routes().
		Add().
		SetName("drxBgpv4PeerRrV6").
		SetNextHopIpv6Address(tc["rxNextHopV6"].(string)).
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)

	drxBgpv4PeerRrV6.Addresses().Add().
		SetAddress(tc["rxAdvRouteV6"].(string)).
		SetPrefix(128).
		SetCount(tc["rxRouteCount"].(uint32)).
		SetStep(1)

	drxBgpv4PeerRrV6.Advanced().
		SetMultiExitDiscriminator(50).
		SetLocalPreference(100).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	drxBgpv4PeerRrV6.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func ibgpRoutePrefixBgpMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	for _, m := range api.GetBgpv4Metrics() {
		if m.SessionState() == gosnappi.Bgpv4MetricSessionState.DOWN ||
			m.RoutesAdvertised() != 2*uint64(tc["txRouteCount"].(uint32)) ||
			m.RoutesReceived() != 2*uint64(tc["rxRouteCount"].(uint32)) {
			return false
		}
	}
	return true
}

func ibgpRoutePrefixBgpPrefixesOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	prefixCount := 0
	for _, m := range api.GetBgpPrefixes() {
		for _, p := range m.Ipv4UnicastPrefixes().Items() {
			for _, key := range []string{"tx", "rx"} {
				if p.Ipv4Address() == tc[key+"AdvRouteV4"].(string) && p.Ipv4NextHop() == tc[key+"NextHopV4"].(string) {
					prefixCount += 1
				}
			}
			if p.LocalPreference() != 100 {
				api.Testing().Logf("Unexpected LocalPref %v \n", p.LocalPreference())
				return false
			}
			if p.MultiExitDiscriminator() != 50 {
				api.Testing().Logf("Unexpected LocalPref %v \n", p.MultiExitDiscriminator())
				return false
			}
		}
		for _, p := range m.Ipv6UnicastPrefixes().Items() {
			for _, key := range []string{"tx", "rx"} {
				if p.Ipv6Address() == tc[key+"AdvRouteV6"].(string) && p.Ipv6NextHop() == tc[key+"NextHopV6"].(string) {
					prefixCount += 1
				}
			}
		}
	}
	return prefixCount == 4
}

func ibgpRoutePrefixFlowMetricsOk(api *otg.OtgApi, tc map[string]interface{}) bool {
	pktCount := uint64(tc["pktCount"].(uint32))

	for _, m := range api.GetFlowMetrics() {
		if m.Transmit() != gosnappi.FlowMetricTransmit.STOPPED ||
			m.FramesTx() != pktCount ||
			m.FramesRx() != pktCount {
			return false
		}

	}

	return true
}
