import logging as log
import pytest
from helpers.otg import otg


def test_ospfv3_p2p_lsa():
    test_const = {
        "pktRate":      50,
        "pktCount":     100,
        "pktSize":      128,
        "txMac":        "00:00:01:01:01:01",
        "txIpV6":       "::1:1:1:1",
        "txGateway":    "::1:1:1:2",
        "txPrefix":     64,
        "rxMac":        "00:00:01:01:01:02",
        "rxIpV6":       "::1:1:1:2",
        "rxGateway":    "::1:1:1:1",
        "rxPrefix":     64,
        "txRouterName": "dtx_ospfv3",
        "rxRouterName": "drx_ospfv3",
        "txRouterId":   "5.5.5.5",
		"rxRouterId":   "7.7.7.7",
        "txRouteCount": 1,
        "rxRouteCount": 1,
        "txAdvRouteV6": "4:4:4:0:0:0:0:1",
		"txAddrPrefix": "4:4:4:0:0:0:0:0",
		"rxAdvRouteV6": "6:6:6:0:0:0:0:1",
		"rxAddrPrefix": "6:6:6:0:0:0:0:0",
		"txMetric": 	10,
		"rxMetric": 	9,
		"txLinkMetric": 20,
		"rxLinkMetric": 19,
    }

    api = otg.OtgApi()
    c = ospfv3_p2p_lsa_config(api, test_const)

    api.set_config(c)

    api.start_protocols()

    api.wait_for(
        fn=lambda: ospfv3_metrics_ok(api, test_const),
        fn_name="wait_for_ospfv3_metrics",
        timeout_seconds=30,
    )

    api.wait_for(
        fn=lambda: ospfv3_lsas_ok(api, test_const),
        fn_name="wait_for_ospfv3_lsas",
        timeout_seconds=30,
    )

    api.start_transmit()

    api.wait_for(
        fn=lambda: flow_metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )


# Please refer to ospfv3 model documentation under 'devices/[ospfv3]' of following url
# for more ospfv3 configuration attributes.
# model: https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/open-traffic-generator/models/master/artifacts/openapi.yaml&nocors#tag/Configuration/operation/set_config


def ospfv3_p2p_lsa_config(api, tc):
    c = api.api.config()
    ptx = c.ports.add(name="ptx", location=api.test_config.otg_ports[0])
    prx = c.ports.add(name="prx", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[ptx.name, prx.name])
    ly.speed = api.test_config.otg_speed

    # transmit
    dtx = c.devices.add(name="dtx")
    drx = c.devices.add(name="drx")

    dtx_eth = dtx.ethernets.add(name="dtx_eth")
    dtx_eth.connection.port_name = ptx.name
    dtx_eth.mac = tc["txMac"]
    dtx_eth.mtu = 1500

    dtx_ipv6 = dtx_eth.ipv6_addresses.add(name="dtx_ipv6")
    dtx_ipv6.address = tc["txIpV6"]
    dtx_ipv6.gateway = tc["txGateway"]
    dtx_ipv6.prefix = tc["txPrefix"]

    dtx_ospfv3 = dtx.ospfv3
    dtx_ospfv3.router_id.custom = tc["txRouterId"]
    
    dtx_ospfv3_instance = dtx_ospfv3.instances.add(name=tc["txRouterName"])
    dtx_ospfv3_instance.store_lsa = True

    dtx_ospfv3_int = dtx_ospfv3_instance.interfaces.add(name="dtx_ospfv3_int")
    dtx_ospfv3_int.ipv6_name = dtx_ipv6.name

    # Note: please change DUT default value for network-type from Broadcast to
    # PointToPoint to make this test interoperable to a port-dut topology
    dtx_ospfv3_int.network_type.choice = dtx_ospfv3_int.network_type.POINT_TO_POINT
    dtx_ospfv3_int.advanced.link_metric = tc["txLinkMetric"]

    dtx_ospfv3_rr6 = dtx_ospfv3_instance.v6_routes.add(name="dtx_ospfv3_rr6")
    dtx_ospfv3_rr6.metric = tc["txMetric"]
    dtx_ospfv3_rr6.addresses.add(
        address=tc["txAdvRouteV6"], prefix=64, count=tc["txRouteCount"], step=1
    )

    # receive
    drx_eth = drx.ethernets.add(name="drx_eth")
    drx_eth.connection.port_name = prx.name
    drx_eth.mac = tc["rxMac"]
    drx_eth.mtu = 1500

    drx_ipv6 = drx_eth.ipv6_addresses.add(name="drx_ipv6")
    drx_ipv6.address = tc["rxIpV6"]
    drx_ipv6.gateway = tc["rxGateway"]
    drx_ipv6.prefix = tc["rxPrefix"]

    drx_ospfv3 = drx.ospfv3
    drx_ospfv3.router_id.custom = tc["rxRouterId"]

    drx_ospfv3_instance = drx_ospfv3.instances.add(name=tc["rxRouterName"])
    drx_ospfv3_instance.store_lsa = True

    drx_ospfv3_int = drx_ospfv3_instance.interfaces.add(name="drx_ospfv3_int")
    drx_ospfv3_int.ipv6_name = drx_ipv6.name

    # Note: please change DUT default value for network-type from Broadcast to
    # PointToPoint to make this test interoperable to a port-dut topology
    drx_ospfv3_int.network_type.choice = drx_ospfv3_int.network_type.POINT_TO_POINT
    drx_ospfv3_int.advanced.link_metric = tc["rxLinkMetric"]

    drx_ospfv3_rr6 = drx_ospfv3_instance.v6_routes.add(name="drx_ospfv3_rr6")
    drx_ospfv3_rr6.metric = tc["rxMetric"]
    drx_ospfv3_rr6.addresses.add(
        address=tc["rxAdvRouteV6"], prefix=64, count=tc["rxRouteCount"], step=1
    )

    # traffic
    for i in range(0, 2):
        f = c.flows.add()
        f.duration.fixed_packets.packets = tc["pktCount"]
        f.rate.pps = tc["pktRate"]
        f.size.fixed = tc["pktSize"]
        f.metrics.enable = True

    ftx_v6 = c.flows[0]
    ftx_v6.name = "ftx_v6"
    ftx_v6.tx_rx.device.tx_names = [dtx_ospfv3_rr6.name]
    ftx_v6.tx_rx.device.rx_names = [drx_ospfv3_rr6.name]

    ftx_v6_eth, ftx_v6_ipv6, ftx_v6_tcp = ftx_v6.packet.ethernet().ipv6().tcp()
    ftx_v6_eth.src.value = dtx_eth.mac
    ftx_v6_ipv6.src.value = tc["txAdvRouteV6"]
    ftx_v6_ipv6.dst.value = tc["rxAdvRouteV6"]
    ftx_v6_tcp.src_port.value = 5000
    ftx_v6_tcp.dst_port.value = 6000

    frx_v6 = c.flows[1]
    frx_v6.name = "frx_v6"
    frx_v6.tx_rx.device.tx_names = [drx_ospfv3_rr6.name]
    frx_v6.tx_rx.device.rx_names = [dtx_ospfv3_rr6.name]

    frx_v6_eth, frx_v6_ip, frx_v6_tcp = frx_v6.packet.ethernet().ipv6().tcp()
    frx_v6_eth.src.value = drx_eth.mac
    frx_v6_ip.src.value = tc["rxAdvRouteV6"]
    frx_v6_ip.dst.value = tc["txAdvRouteV6"]
    frx_v6_tcp.src_port.value = 5000
    frx_v6_tcp.dst_port.value = 6000

    log.info("Config:\n%s", c)
    return c


def ospfv3_metrics_ok(api, tc):
    for m in api.get_ospfv3_metrics():
        if m.full_state_count < 1 or m.lsa_sent < 3 or m.lsa_received < 3:
            return False
    return True


def ospfv3_lsas_ok(api, tc):
    lsa_count = 0

    adv_router_id = ""
    addr_prefix = ""
    metric = 0
    router_lsa_nbr_id = ""
    router_lsa_link_metric = 0
    router_lsa_link_type = "point_to_point"

    for m in api.get_ospfv3_lsas():
        if m.router_name == tc["txRouterName"]:
            adv_router_id = tc["rxRouterId"]
            addr_prefix = tc["rxAddrPrefix"]
            router_lsa_nbr_id = tc["txRouterId"]
            metric = tc["rxMetric"]
            router_lsa_link_metric = tc["rxLinkMetric"]

        if m.router_name == tc["rxRouterName"]:
            adv_router_id = tc["txRouterId"]
            addr_prefix = tc["txAddrPrefix"]
            router_lsa_nbr_id = tc["rxRouterId"]
            metric = tc["txMetric"]
            router_lsa_link_metric = tc["txLinkMetric"]

        # validate lsas
        if (
            len(m.inter_area_prefix_lsas) == 1
            and m.inter_area_prefix_lsas[0].address_prefix == addr_prefix
            and m.inter_area_prefix_lsas[0].header.advertising_router_id == adv_router_id
            and m.inter_area_prefix_lsas[0].metric == metric
        ):
            lsa_count += 1

        if (
            len(m.link_lsas) == 1
            and m.link_lsas[0].header.advertising_router_id == adv_router_id
        ):
            lsa_count += 1

        if (
            len(m.router_lsas) == 1
            and m.router_lsas[0].header.advertising_router_id == adv_router_id
            and m.router_lsas[0].neighbor_router_id == router_lsa_nbr_id
            and len(m.router_lsas[0].links) == 1
            and m.router_lsas[0].links[0].type == router_lsa_link_type
            and m.router_lsas[0].links[0].metric == router_lsa_link_metric
        ):
            lsa_count += 1

    return lsa_count == 6


def flow_metrics_ok(api, tc):
    for m in api.get_flow_metrics():
        if (
            m.transmit != m.STOPPED
            or m.frames_tx != tc["pktCount"]
            or m.frames_rx != tc["pktCount"]
        ):
            return False
    return True
