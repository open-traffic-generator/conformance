import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.all
@pytest.mark.feature
@pytest.mark.b2b
def test_isis_route_install():
    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "txIp": "1.1.1.1",
        "txGateway": "1.1.1.2",
        "txPrefix": 24,
        "txIpv6": "1100::1",
        "txv6Gateway": "1100::2",
        "txv6Prefix": 64,
        "rxMac": "00:00:01:01:01:02",
        "rxIp": "1.1.1.2",
        "rxGateway": "1.1.1.1",
        "rxPrefix": 24,
        "rxIpv6": "1100::2",
        "rxv6Gateway": "1100::1",
        "rxv6Prefix": 64,
        "txRouteCount": 1,
        "rxRouteCount": 1,
        "txAdvRouteV4": "10.10.10.1",
        "rxAdvRouteV4": "20.20.20.1",
        "txAdvRouteV6": "::10:10:10:01",
        "rxAdvRouteV6": "::20:20:20:01",
    }

    api = otg.OtgApi()
    c = isis_route_install_config(api, test_const)

    api.set_config(c)

    api.start_protocols()

    api.wait_for(
        fn=lambda: isis_metrics_ok(api, test_const),
        fn_name="wait_for_isis_metrics",
        timeout_seconds=20,
    )

    api.start_transmit()

    api.wait_for(
        fn=lambda: flow_metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )


def isis_route_install_config(api, tc):
    c = api.api.config()
    ptx = c.ports.add(name="ptx", location=api.test_config.otg_ports[0])
    prx = c.ports.add(name="prx", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[ptx.name, prx.name])
    ly.speed = api.test_config.otg_speed

    # transmitter
    tx = c.devices.add(name="tx")
    rx = c.devices.add(name="rx")

    tx_eth = tx.ethernets.add(name="tx_eth")
    tx_eth.port_name = ptx.name
    tx_eth.mac = tc["txMac"]
    tx_eth.mtu = 1500

    tx_ip = tx_eth.ipv4_addresses.add(name="tx_ip")
    tx_ip.address = tc["txIp"]
    tx_ip.gateway = tc["txGateway"]
    tx_ip.prefix = tc["txPrefix"]

    tx_ipv6 = tx_eth.ipv6_addresses.add(name="txv6_ip")
    tx_ipv6.address = tc["txIpv6"]
    tx_ipv6.gateway = tc["txv6Gateway"]
    tx_ipv6.prefix = tc["txv6Prefix"]

    tx.isis.system_id = "640000000001"
    tx.isis.name = "tx_isis"

    tx.isis.advanced.area_addresses = ["490001"]
    tx.isis.advanced.lsp_refresh_rate = 900
    tx.isis.advanced.enable_attached_bit = False

    tx.isis.basic.ipv4_te_router_id = tc["txIp"]
    tx.isis.basic.hostname = "ixia-c-port1"

    tx_isis_int = tx.isis.interfaces.add()
    tx_isis_int.eth_name = tx_eth.name
    tx_isis_int.name = "tx_isis_int"
    tx_isis_int.network_type = tx_isis_int.POINT_TO_POINT
    tx_isis_int.level_type = "level_1_2"

    tx_isis_int.l2_settings.dead_interval = 30
    tx_isis_int.l2_settings.hello_interval = 10
    tx_isis_int.l2_settings.priority = 0

    tx_isis_int.advanced.auto_adjust_supported_protocols = True

    tx_isis_rr4 = tx.isis.v4_routes.add(name="tx_isis_rr4")
    tx_isis_rr4.link_metric = 10
    tx_isis_rr4.addresses.add(
        address=tc["txAdvRouteV4"], prefix=32, count=tc["txRouteCount"], step=1
    )

    tx_isis_rrv6 = tx.isis.v6_routes.add(name="tx_isis_rr6")
    tx_isis_rrv6.addresses.add(
        address=tc["txAdvRouteV6"], prefix=32, count=tc["txRouteCount"], step=1
    )

    # receiver
    rx_eth = rx.ethernets.add(name="rx_eth")
    rx_eth.port_name = prx.name
    rx_eth.mac = tc["rxMac"]
    rx_eth.mtu = 1500

    rx_ip = rx_eth.ipv4_addresses.add(name="rx_ip")
    rx_ip.address = tc["rxIp"]
    rx_ip.gateway = tc["rxGateway"]
    rx_ip.prefix = tc["rxPrefix"]

    rx_ipv6 = rx_eth.ipv6_addresses.add(name="rxv6_ip")
    rx_ipv6.address = tc["rxIpv6"]
    rx_ipv6.gateway = tc["rxv6Gateway"]
    rx_ipv6.prefix = tc["rxv6Prefix"]

    rx.isis.system_id = "640000000001"  # "650000000001"
    rx.isis.name = "rx_isis"

    rx.isis.advanced.area_addresses = ["490001"]
    rx.isis.advanced.lsp_refresh_rate = 900
    rx.isis.advanced.enable_attached_bit = False

    rx.isis.basic.ipv4_te_router_id = tc["rxIp"]
    rx.isis.basic.hostname = "ixia-c-port2"

    rx_isis_int = rx.isis.interfaces.add()
    rx_isis_int.eth_name = rx_eth.name
    rx_isis_int.name = "rx_isis_int"
    rx_isis_int.network_type = rx_isis_int.POINT_TO_POINT
    rx_isis_int.level_type = "level_1_2"

    rx_isis_int.l2_settings.dead_interval = 30
    rx_isis_int.l2_settings.hello_interval = 10
    rx_isis_int.l2_settings.priority = 0

    rx_isis_int.advanced.auto_adjust_supported_protocols = True

    rx_isis_rr4 = rx.isis.v4_routes.add(name="rx_isis_rr4")
    rx_isis_rr4.link_metric = 10
    rx_isis_rr4.addresses.add(
        address=tc["rxAdvRouteV4"], prefix=32, count=tc["rxRouteCount"], step=1
    )

    rx_isis_rrv6 = rx.isis.v6_routes.add(name="rx_isis_rr6")
    rx_isis_rrv6.addresses.add(
        address=tc["rxAdvRouteV6"], prefix=32, count=tc["rxRouteCount"], step=1
    )

    for i in range(0, 4):
        f = c.flows.add()
        f.duration.fixed_packets.packets = tc["pktCount"]
        f.rate.pps = tc["pktRate"]
        f.size.fixed = tc["pktSize"]
        f.metrics.enable = True

    ftx_v4 = c.flows[0]
    ftx_v4.name = "ftx_v4"
    ftx_v4.tx_rx.device.tx_names = [tx_isis_rr4.name]
    ftx_v4.tx_rx.device.rx_names = [rx_isis_rr4.name]

    ftx_v4_eth, ftx_v4_ip, ftx_v4_tcp = ftx_v4.packet.ethernet().ipv4().tcp()
    ftx_v4_eth.src.value = tx_eth.mac
    ftx_v4_ip.src.value = tc["txAdvRouteV4"]
    ftx_v4_ip.dst.value = tc["rxAdvRouteV4"]
    ftx_v4_tcp.src_port.value = 5000
    ftx_v4_tcp.dst_port.value = 6000

    ftx_v6 = c.flows[1]
    ftx_v6.name = "ftx_v6"
    ftx_v6.tx_rx.device.tx_names = [tx_isis_rrv6.name]
    ftx_v6.tx_rx.device.rx_names = [rx_isis_rrv6.name]

    ftx_v6_eth, ftx_v6_ip, ftx_v6_tcp = ftx_v6.packet.ethernet().ipv6().tcp()
    ftx_v6_eth.src.value = tx_eth.mac
    ftx_v6_ip.src.value = tc["txAdvRouteV6"]
    ftx_v6_ip.dst.value = tc["rxAdvRouteV6"]
    ftx_v6_tcp.src_port.value = 5000
    ftx_v6_tcp.dst_port.value = 6000

    frx_v4 = c.flows[2]
    frx_v4.name = "frx_v4"
    frx_v4.tx_rx.device.tx_names = [rx_isis_rr4.name]
    frx_v4.tx_rx.device.rx_names = [tx_isis_rr4.name]

    frx_v4_eth, frx_v4_ip, frx_v4_tcp = frx_v4.packet.ethernet().ipv4().tcp()
    frx_v4_eth.src.value = rx_eth.mac
    frx_v4_ip.src.value = tc["rxAdvRouteV4"]
    frx_v4_ip.dst.value = tc["txAdvRouteV4"]
    frx_v4_tcp.src_port.value = 5000
    frx_v4_tcp.dst_port.value = 6000

    frx_v6 = c.flows[3]
    frx_v6.name = "frx_v6"
    frx_v6.tx_rx.device.tx_names = [rx_isis_rrv6.name]
    frx_v6.tx_rx.device.rx_names = [tx_isis_rrv6.name]

    frx_v6_eth, frx_v6_ip, frx_v6_tcp = frx_v6.packet.ethernet().ipv6().tcp()
    frx_v6_eth.src.value = rx_eth.mac
    frx_v6_ip.src.value = tc["rxAdvRouteV6"]
    frx_v6_ip.dst.value = tc["txAdvRouteV6"]
    frx_v6_tcp.src_port.value = 5000
    frx_v6_tcp.dst_port.value = 6000

    log.info("Config:\n%s", c)
    return c


def isis_metrics_ok(api, tc):
    for m in api.get_isis_metrics():
        if (
            m.l1_sessions_up != 1
            or m.l1_database_size != 1
            or m.l2_sessions_up != 1
            or m.l2_database_size != 1
        ):
            return False
    return True


def flow_metrics_ok(api, tc):
    for m in api.get_flow_metrics():
        if (
            m.transmit != m.STOPPED
            or m.frames_tx != tc["pktCount"]
            or m.frames_rx != tc["pktCount"]
        ):
            return False
    return True
