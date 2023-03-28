import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.all
@pytest.mark.cpdp
def test_ip_neighbors():
    # TODO: add support for IPv6 as well
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
    }

    api = otg.OtgApi()
    c = ip_neighbors_config(api, test_const)

    api.set_config(c)

    api.wait_for(
        fn=lambda: ipv4_neighbors_ok(api, test_const), fn_name="wait_for_ipv4_neighbors"
    )

    api.start_transmit()

    api.wait_for(
        fn=lambda: flow_metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )


def ip_neighbors_config(api, tc):
    c = api.api.config()
    ptx = c.ports.add(name="ptx", location=api.test_config.otg_ports[0])
    prx = c.ports.add(name="prx", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[ptx.name, prx.name])
    ly.speed = api.test_config.otg_speed

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

    drx_eth = drx.ethernets.add(name="drx_eth")
    drx_eth.connection.port_name = prx.name
    drx_eth.mac = tc["rxMac"]
    drx_eth.mtu = 1500

    drx_ip = drx_eth.ipv4_addresses.add(name="drx_ip")
    drx_ip.address = tc["rxIp"]
    drx_ip.gateway = tc["rxGateway"]
    drx_ip.prefix = tc["rxPrefix"]

    flow = c.flows.add()
    flow.name = "ftx_v4"
    flow.duration.fixed_packets.packets = tc["pktCount"]
    flow.rate.pps = tc["pktRate"]
    flow.size.fixed = tc["pktSize"]
    flow.metrics.enable = True

    flow.tx_rx.device.tx_names = [dtx_ip.name]
    flow.tx_rx.device.rx_names = [drx_ip.name]

    ftx_v4_eth, ftx_v4_ip = flow.packet.ethernet().ipv4()
    ftx_v4_eth.src.value = dtx_eth.mac
    ftx_v4_ip.src.value = tc["txIp"]
    ftx_v4_ip.dst.value = tc["rxIp"]

    log.info("Config:\n%s", c)
    return c


def ipv4_neighbors_ok(api, tc):
    count = 0
    for n in api.get_ipv4_neighbors():
        if n.link_layer_address is not None:
            for key in ["txGateway", "rxGateway"]:
                if n.ipv4_address == tc[key]:
                    count += 1

    return count == 2


def flow_metrics_ok(api, tc):
    for m in api.get_flow_metrics():
        if (
            m.transmit != m.STOPPED
            or m.frames_tx != tc["pktCount"]
            or m.frames_rx != tc["pktCount"]
        ):
            return False
    return True
