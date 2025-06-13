import logging as log
import pytest
from helpers.otg import otg


def test_ospfv2_p2p_lsa():
    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "txIp": "1.1.1.1",
        "txGateway": "1.1.1.2",
        "txPrefix": 24,
        "rxMac": "00:00:01:01:01:02",
        "rxIp": "1.1.1.2",
        "rxGateway": "1.1.1.1",
        "rxPrefix": 24,
        "txRouterName": "dtx_ospfv2",
        "rxRouterName": "drx_ospfv2",
        "txRouteCount": 1,
        "rxRouteCount": 1,
        "txAdvRouteV4": "10.10.10.1",
        "rxAdvRouteV4": "20.20.20.1",
    }

    api = otg.OtgApi()
    c = ospfv2_p2p_lsa_config(api, test_const)
    api.api.request_timeout = 300
    api.set_config(c)

    api.start_protocols()

    api.wait_for(
        fn=lambda: ospfv2_metrics_ok(api, test_const),
        fn_name="wait_for_ospfv2_metrics",
        timeout_seconds=30,
    )

    api.wait_for(
        fn=lambda: ospfv2_lsas_ok(api, test_const),
        fn_name="wait_for_ospfv2_lsas",
        timeout_seconds=30,
    )

    api.start_transmit()

    api.wait_for(
        fn=lambda: flow_metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )


# Please refer to ospfv2 model documentation under 'devices/[ospfv2]' of following url
# for more ospfv2 configuration attributes.
# model: https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/open-traffic-generator/models/master/artifacts/openapi.yaml&nocors#tag/Configuration/operation/set_config


def ospfv2_p2p_lsa_config(api, tc):
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

    dtx_ip = dtx_eth.ipv4_addresses.add(name="dtx_ip")
    dtx_ip.address = tc["txIp"]
    dtx_ip.gateway = tc["txGateway"]
    dtx_ip.prefix = tc["txPrefix"]

    dtx.ospfv2.name = tc["txRouterName"]
    dtx.ospfv2.store_lsa = True

    dtx_ospfv2_int = dtx.ospfv2.interfaces.add(name="dtx_ospfv2_int")
    dtx_ospfv2_int.ipv4_name = dtx_ip.name

    # Note: please change DUT default value for network-type from Broadcast to
    # PointToPoint to make this test interoperable to a port-dut topology
    dtx_ospfv2_int.network_type.choice = dtx_ospfv2_int.network_type.POINT_TO_POINT

    dtx_ospfv2_rr4 = dtx.ospfv2.v4_routes.add(name="dtx_ospfv2_rr4")
    dtx_ospfv2_rr4.metric = 10
    dtx_ospfv2_rr4.addresses.add(
        address=tc["txAdvRouteV4"], prefix=32, count=tc["txRouteCount"], step=1
    )

    # receive
    drx_eth = drx.ethernets.add(name="drx_eth")
    drx_eth.connection.port_name = prx.name
    drx_eth.mac = tc["rxMac"]
    drx_eth.mtu = 1500

    drx_ip = drx_eth.ipv4_addresses.add(name="drx_ip")
    drx_ip.address = tc["rxIp"]
    drx_ip.gateway = tc["rxGateway"]
    drx_ip.prefix = tc["rxPrefix"]

    drx.ospfv2.name = tc["rxRouterName"]
    drx.ospfv2.store_lsa = True

    drx_ospfv2_int = drx.ospfv2.interfaces.add(name="drx_ospfv2_int")
    drx_ospfv2_int.ipv4_name = drx_ip.name

    # Note: please change DUT default value for network-type from Broadcast to
    # PointToPoint to make this test interoperable to a port-dut topology
    drx_ospfv2_int.network_type.choice = drx_ospfv2_int.network_type.POINT_TO_POINT

    drx_ospfv2_rr4 = drx.ospfv2.v4_routes.add(name="drx_ospfv2_rr4")
    drx_ospfv2_rr4.metric = 10
    drx_ospfv2_rr4.addresses.add(
        address=tc["rxAdvRouteV4"], prefix=32, count=tc["rxRouteCount"], step=1
    )

    # traffic
    for i in range(0, 2):
        f = c.flows.add()
        f.duration.fixed_packets.packets = tc["pktCount"]
        f.rate.pps = tc["pktRate"]
        f.size.fixed = tc["pktSize"]
        f.metrics.enable = True

    ftx_v4 = c.flows[0]
    ftx_v4.name = "ftx_v4"
    ftx_v4.tx_rx.device.tx_names = [dtx_ospfv2_rr4.name]
    ftx_v4.tx_rx.device.rx_names = [drx_ospfv2_rr4.name]

    ftx_v4_eth, ftx_v4_ip, ftx_v4_tcp = ftx_v4.packet.ethernet().ipv4().tcp()
    ftx_v4_eth.src.value = dtx_eth.mac
    ftx_v4_ip.src.value = tc["txAdvRouteV4"]
    ftx_v4_ip.dst.value = tc["rxAdvRouteV4"]
    ftx_v4_tcp.src_port.value = 5000
    ftx_v4_tcp.dst_port.value = 6000

    frx_v4 = c.flows[1]
    frx_v4.name = "frx_v4"
    frx_v4.tx_rx.device.tx_names = [drx_ospfv2_rr4.name]
    frx_v4.tx_rx.device.rx_names = [dtx_ospfv2_rr4.name]

    frx_v4_eth, frx_v4_ip, frx_v4_tcp = frx_v4.packet.ethernet().ipv4().tcp()
    frx_v4_eth.src.value = drx_eth.mac
    frx_v4_ip.src.value = tc["rxAdvRouteV4"]
    frx_v4_ip.dst.value = tc["txAdvRouteV4"]
    frx_v4_tcp.src_port.value = 5000
    frx_v4_tcp.dst_port.value = 6000

    log.info("Config:\n%s", c)
    return c


def ospfv2_metrics_ok(api, tc):
    for m in api.get_ospfv2_metrics():
        if m.full_state_count < 1 or m.lsa_sent < 2 or m.lsa_received < 2:
            return False
    return True


def ospfv2_lsas_ok(api, tc):
    lsa_count = 0

    adv_router_id = ""
    nw_summary_lsa_id = ""
    router_lsa_id = ""
    router_lsa_link_id = ""
    router_lsa_link_data = ""

    for m in api.get_ospfv2_lsas():
        if m.router_name == tc["txRouterName"]:
            adv_router_id = tc["rxIp"]
            nw_summary_lsa_id = tc["rxAdvRouteV4"]
            router_lsa_id = tc["rxIp"]
            router_lsa_link_id = tc["txIp"]
            router_lsa_link_data = tc["rxIp"]

        if m.router_name == tc["rxRouterName"]:
            adv_router_id = tc["txIp"]
            nw_summary_lsa_id = tc["txAdvRouteV4"]
            router_lsa_id = tc["txIp"]
            router_lsa_link_id = tc["rxIp"]
            router_lsa_link_data = tc["txIp"]

        # validate lsas
        if (
            len(m.network_summary_lsas) == 1
            and m.network_summary_lsas[0].metric == 10
            and m.network_summary_lsas[0].header.advertising_router_id == adv_router_id
            and m.network_summary_lsas[0].header.lsa_id == nw_summary_lsa_id
        ):
            lsa_count += 1
        if (
            len(m.router_lsas) == 1
            and m.router_lsas[0].header.advertising_router_id == adv_router_id
            and m.router_lsas[0].header.lsa_id == router_lsa_id
            and len(m.router_lsas[0].links) == 2
        ):
            has_stub = False
            has_p2p = False
            for link in (m.router_lsas[0].links):
                if (link.type == "stub" and link.metric == 0 or 1):
                    has_stub = True
                if (link.type == "point_to_point"
                    and link.id == router_lsa_link_id
                    and link.data == router_lsa_link_data
                    and link.metric == 0 or 1):
                    has_p2p = True
 
            if has_stub and has_p2p:
                lsa_count += 1
                
    return lsa_count == 4


def flow_metrics_ok(api, tc):
    for m in api.get_flow_metrics():
        if (
            m.transmit != m.STOPPED
            or m.frames_tx != tc["pktCount"]
            or m.frames_rx != tc["pktCount"]
        ):
            return False
    return True
