import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.all
@pytest.mark.cpdp
def test_ebgp_route_prefix():
    # TODO: add support for BGP for IPv6 as well
    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "txIp": "1.1.1.1",
        "txGateway": "1.1.1.2",
        "txPrefix": 24,
        "txAs": 1111,
        "rxMac": "00:00:01:01:01:02",
        "rxIp": "1.1.1.2",
        "rxGateway": "1.1.1.1",
        "rxPrefix": 4,
        "rxAs": 1112,
        "txRouteCount": 1,
        "rxRouteCount": 1,
        "txNextHopV4": "1.1.1.3",
        "txNextHopV6": "::1:1:1:3",
        "rxNextHopV4": "1.1.1.4",
        "rxNextHopV6": "::1:1:1:4",
        "txAdvRouteV4": "10.10.10.1",
        "rxAdvRouteV4": "20.20.20.1",
        "txAdvRouteV6": "::10:10:10:1",
        "rxAdvRouteV6": "::20:20:20:1",
    }

    api = otg.OtgApi()
    c = ebgp_route_prefix_config(api, test_const)

    api.set_config(c)

    api.start_protocols()

    api.wait_for(
        fn=lambda: bgp_metrics_ok(api, test_const),
        fn_name="wait_for_bgp_metrics",
    )

    api.wait_for(
        fn=lambda: bgp_prefixes_ok(api, test_const),
        fn_name="wait_for_bgp_prefixes",
    )

    api.start_transmit()

    api.wait_for(
        fn=lambda: flow_metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )


def ebgp_route_prefix_config(api, tc):
    c = api.api.config()
    ptx = c.ports.add(name="ptx", location=api.test_config.otg_ports[0])
    prx = c.ports.add(name="prx", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[ptx.name, prx.name])
    ly.speed = api.test_config.otg_speed

    dtx = c.devices.add(name="dtx")
    drx = c.devices.add(name="drx")

    dtx_eth = dtx.ethernets.add(name="dtx_eth")
    dtx_eth.port_name = ptx.name
    dtx_eth.mac = tc["txMac"]
    dtx_eth.mtu = 1500

    dtx_ip = dtx_eth.ipv4_addresses.add(name="dtx_ip")
    dtx_ip.address = tc["txIp"]
    dtx_ip.gateway = tc["txGateway"]
    dtx_ip.prefix = tc["txPrefix"]

    dtx.bgp.router_id = tc["txIp"]

    dtx_bgpv4 = dtx.bgp.ipv4_interfaces.add()
    dtx_bgpv4.ipv4_name = dtx_ip.name

    dtx_bgpv4_peer = dtx_bgpv4.peers.add(name="dtx_bgpv4_peer")
    dtx_bgpv4_peer.as_number = tc["txAs"]
    dtx_bgpv4_peer.as_type = dtx_bgpv4_peer.EBGP
    dtx_bgpv4_peer.peer_address = tc["txGateway"]
    dtx_bgpv4_peer.learned_information_filter.unicast_ipv4_prefix = True
    dtx_bgpv4_peer.learned_information_filter.unicast_ipv6_prefix = True

    dtx_bgpv4_peer_rrv4 = dtx_bgpv4_peer.v4_routes.add(name="dtx_bgpv4_peer_rrv4")
    dtx_bgpv4_peer_rrv4.next_hop_ipv4_address = tc["txNextHopV4"]
    dtx_bgpv4_peer_rrv4.next_hop_address_type = dtx_bgpv4_peer_rrv4.IPV4
    dtx_bgpv4_peer_rrv4.next_hop_mode = dtx_bgpv4_peer_rrv4.MANUAL

    dtx_bgpv4_peer_rrv4.addresses.add(
        address=tc["txAdvRouteV4"], prefix=32, count=tc["txRouteCount"], step=1
    )

    dtx_bgpv4_peer_rrv4.advanced.multi_exit_discriminator = 50
    dtx_bgpv4_peer_rrv4.advanced.origin = dtx_bgpv4_peer_rrv4.advanced.EGP

    dtx_bgpv4_peer_rrv4_com = dtx_bgpv4_peer_rrv4.communities.add(
        as_number=1, as_custom=2
    )
    dtx_bgpv4_peer_rrv4_com.type = dtx_bgpv4_peer_rrv4_com.MANUAL_AS_NUMBER

    dtx_bgpv4_peer_rrv4.as_path.as_set_mode = dtx_bgpv4_peer_rrv4.as_path.INCLUDE_AS_SET

    dtx_bgpv4_peer_rrv4_seg = dtx_bgpv4_peer_rrv4.as_path.segments.add()
    dtx_bgpv4_peer_rrv4_seg.as_numbers = [1112, 1113]
    dtx_bgpv4_peer_rrv4_seg.type = dtx_bgpv4_peer_rrv4_seg.AS_SEQ

    dtx_bgpv4_peer_rrv6 = dtx_bgpv4_peer.v6_routes.add(name="dtx_bgpv4_peer_rrv6")
    dtx_bgpv4_peer_rrv6.next_hop_ipv6_address = tc["txNextHopV6"]
    dtx_bgpv4_peer_rrv6.next_hop_address_type = dtx_bgpv4_peer_rrv6.IPV6
    dtx_bgpv4_peer_rrv6.next_hop_mode = dtx_bgpv4_peer_rrv6.MANUAL

    dtx_bgpv4_peer_rrv6.addresses.add(
        address=tc["txAdvRouteV6"], prefix=128, count=tc["txRouteCount"], step=1
    )

    dtx_bgpv4_peer_rrv6.advanced.multi_exit_discriminator = 50
    dtx_bgpv4_peer_rrv6.advanced.origin = dtx_bgpv4_peer_rrv6.advanced.EGP

    dtx_bgpv4_peer_rrv6_com = dtx_bgpv4_peer_rrv6.communities.add(
        as_number=1, as_custom=2
    )
    dtx_bgpv4_peer_rrv6_com.type = dtx_bgpv4_peer_rrv6_com.MANUAL_AS_NUMBER

    dtx_bgpv4_peer_rrv6.as_path.as_set_mode = dtx_bgpv4_peer_rrv6.as_path.INCLUDE_AS_SET

    dtx_bgpv4_peer_rrv6_seg = dtx_bgpv4_peer_rrv6.as_path.segments.add()
    dtx_bgpv4_peer_rrv6_seg.as_numbers = [1112, 1113]
    dtx_bgpv4_peer_rrv6_seg.type = dtx_bgpv4_peer_rrv6_seg.AS_SEQ

    drx_eth = drx.ethernets.add(name="drx_eth")
    drx_eth.port_name = prx.name
    drx_eth.mac = tc["rxMac"]
    drx_eth.mtu = 1500

    drx_ip = drx_eth.ipv4_addresses.add(name="drx_ip")
    drx_ip.address = tc["rxIp"]
    drx_ip.gateway = tc["rxGateway"]
    drx_ip.prefix = tc["rxPrefix"]

    drx.bgp.router_id = tc["rxIp"]

    drx_bgpv4 = drx.bgp.ipv4_interfaces.add()
    drx_bgpv4.ipv4_name = drx_ip.name

    drx_bgpv4_peer = drx_bgpv4.peers.add(name="drx_bgpv4_peer")
    drx_bgpv4_peer.as_number = tc["rxAs"]
    drx_bgpv4_peer.as_type = drx_bgpv4_peer.EBGP
    drx_bgpv4_peer.peer_address = tc["rxGateway"]
    drx_bgpv4_peer.learned_information_filter.unicast_ipv4_prefix = True
    drx_bgpv4_peer.learned_information_filter.unicast_ipv6_prefix = True

    drx_bgpv4_peer_rrv4 = drx_bgpv4_peer.v4_routes.add(name="drx_bgpv4_peer_rrv4")
    drx_bgpv4_peer_rrv4.next_hop_ipv4_address = tc["rxNextHopV4"]
    drx_bgpv4_peer_rrv4.next_hop_address_type = drx_bgpv4_peer_rrv4.IPV4
    drx_bgpv4_peer_rrv4.next_hop_mode = drx_bgpv4_peer_rrv4.MANUAL

    drx_bgpv4_peer_rrv4.addresses.add(
        address=tc["rxAdvRouteV4"], prefix=32, count=tc["rxRouteCount"], step=1
    )

    drx_bgpv4_peer_rrv4.advanced.multi_exit_discriminator = 50
    drx_bgpv4_peer_rrv4.advanced.origin = drx_bgpv4_peer_rrv4.advanced.EGP

    drx_bgpv4_peer_rrv4_com = drx_bgpv4_peer_rrv4.communities.add(
        as_number=1, as_custom=2
    )
    drx_bgpv4_peer_rrv4_com.type = drx_bgpv4_peer_rrv4_com.MANUAL_AS_NUMBER

    drx_bgpv4_peer_rrv4.as_path.as_set_mode = drx_bgpv4_peer_rrv4.as_path.INCLUDE_AS_SET

    drx_bgpv4_peer_rrv4_seg = drx_bgpv4_peer_rrv4.as_path.segments.add()
    drx_bgpv4_peer_rrv4_seg.as_numbers = [1112, 1113]
    drx_bgpv4_peer_rrv4_seg.type = drx_bgpv4_peer_rrv4_seg.AS_SEQ

    drx_bgpv4_peer_rrv6 = drx_bgpv4_peer.v6_routes.add(name="drx_bgpv4_peer_rrv6")
    drx_bgpv4_peer_rrv6.next_hop_ipv6_address = tc["rxNextHopV6"]
    drx_bgpv4_peer_rrv6.next_hop_address_type = drx_bgpv4_peer_rrv6.IPV6
    drx_bgpv4_peer_rrv6.next_hop_mode = drx_bgpv4_peer_rrv6.MANUAL

    drx_bgpv4_peer_rrv6.addresses.add(
        address=tc["rxAdvRouteV6"], prefix=128, count=tc["rxRouteCount"], step=1
    )

    drx_bgpv4_peer_rrv6.advanced.multi_exit_discriminator = 50
    drx_bgpv4_peer_rrv6.advanced.origin = drx_bgpv4_peer_rrv6.advanced.EGP

    drx_bgpv4_peer_rrv6_com = drx_bgpv4_peer_rrv6.communities.add(
        as_number=1, as_custom=2
    )
    drx_bgpv4_peer_rrv6_com.type = drx_bgpv4_peer_rrv6_com.MANUAL_AS_NUMBER

    drx_bgpv4_peer_rrv6.as_path.as_set_mode = drx_bgpv4_peer_rrv6.as_path.INCLUDE_AS_SET

    drx_bgpv4_peer_rrv6_seg = drx_bgpv4_peer_rrv6.as_path.segments.add()
    drx_bgpv4_peer_rrv6_seg.as_numbers = [1112, 1113]
    drx_bgpv4_peer_rrv6_seg.type = drx_bgpv4_peer_rrv6_seg.AS_SEQ

    for i in range(0, 4):
        f = c.flows.add()
        f.duration.fixed_packets.packets = tc["pktCount"]
        f.rate.pps = tc["pktRate"]
        f.size.fixed = tc["pktSize"]
        f.metrics.enable = True

    ftx_v4 = c.flows[0]
    ftx_v4.name = "ftx_v4"
    ftx_v4.tx_rx.device.tx_names = [dtx_bgpv4_peer_rrv4.name]
    ftx_v4.tx_rx.device.rx_names = [drx_bgpv4_peer_rrv4.name]

    ftx_v4_eth, ftx_v4_ip, ftx_v4_tcp = ftx_v4.packet.ethernet().ipv4().tcp()
    ftx_v4_eth.src.value = dtx_eth.mac
    ftx_v4_ip.src.value = tc["txAdvRouteV4"]
    ftx_v4_ip.dst.value = tc["rxAdvRouteV4"]
    ftx_v4_tcp.src_port.value = 5000
    ftx_v4_tcp.dst_port.value = 6000

    ftx_v6 = c.flows[1]
    ftx_v6.name = "ftx_v6"
    ftx_v6.tx_rx.device.tx_names = [dtx_bgpv4_peer_rrv6.name]
    ftx_v6.tx_rx.device.rx_names = [drx_bgpv4_peer_rrv6.name]

    ftx_v6_eth, ftx_v6_ip, ftx_v6_tcp = ftx_v6.packet.ethernet().ipv6().tcp()
    ftx_v6_eth.src.value = dtx_eth.mac
    ftx_v6_ip.src.value = tc["txAdvRouteV6"]
    ftx_v6_ip.dst.value = tc["rxAdvRouteV6"]
    ftx_v6_tcp.src_port.value = 5000
    ftx_v6_tcp.dst_port.value = 6000

    frx_v4 = c.flows[2]
    frx_v4.name = "frx_v4"
    frx_v4.tx_rx.device.tx_names = [drx_bgpv4_peer_rrv4.name]
    frx_v4.tx_rx.device.rx_names = [dtx_bgpv4_peer_rrv4.name]

    frx_v4_eth, frx_v4_ip, frx_v4_tcp = frx_v4.packet.ethernet().ipv4().tcp()
    frx_v4_eth.src.value = drx_eth.mac
    frx_v4_ip.src.value = tc["rxAdvRouteV4"]
    frx_v4_ip.dst.value = tc["txAdvRouteV4"]
    frx_v4_tcp.src_port.value = 5000
    frx_v4_tcp.dst_port.value = 6000

    frx_v6 = c.flows[3]
    frx_v6.name = "frx_v6"
    frx_v6.tx_rx.device.tx_names = [drx_bgpv4_peer_rrv6.name]
    frx_v6.tx_rx.device.rx_names = [dtx_bgpv4_peer_rrv6.name]

    frx_v6_eth, frx_v6_ip, frx_v6_tcp = frx_v6.packet.ethernet().ipv6().tcp()
    frx_v6_eth.src.value = drx_eth.mac
    frx_v6_ip.src.value = tc["rxAdvRouteV6"]
    frx_v6_ip.dst.value = tc["txAdvRouteV6"]
    frx_v6_tcp.src_port.value = 5000
    frx_v6_tcp.dst_port.value = 6000

    log.info("Config:\n%s", c)
    return c


def bgp_metrics_ok(api, tc):
    for m in api.get_bgpv4_metrics():
        if (
            m.session_state == m.DOWN
            or m.routes_advertised != 2 * tc["txRouteCount"]
            or m.routes_received != 2 * tc["rxRouteCount"]
        ):
            return False
    return True


def bgp_prefixes_ok(api, tc):
    prefix_count = 0
    for m in api.get_bgp_prefixes():
        for p in m.ipv4_unicast_prefixes:
            for key in ["tx", "rx"]:
                if (
                    p.ipv4_address == tc[key + "AdvRouteV4"]
                    and p.ipv4_next_hop == tc[key + "NextHopV4"]
                ):
                    prefix_count += 1
        for p in m.ipv6_unicast_prefixes:
            for key in ["tx", "rx"]:
                if (
                    p.ipv6_address == tc[key + "AdvRouteV6"]
                    and p.ipv6_next_hop == tc[key + "NextHopV6"]
                ):
                    prefix_count += 1

    return prefix_count == 4


def flow_metrics_ok(api, tc):
    for m in api.get_flow_metrics():
        if (
            m.transmit != m.STOPPED
            or m.frames_tx != tc["pktCount"]
            or m.frames_rx != tc["pktCount"]
        ):
            return False
    return True
