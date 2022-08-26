import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.free
@pytest.mark.b2b
def test_tcp_header_valuelist_ports():
    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "rxMac": "00:00:01:01:01:02",
        "txIp": "1.1.1.1",
        "rxIp": "1.1.1.2",
        "txTcpPortValueList": [5000, 5010, 5020, 5030],
        "rxTcpPortValueList": [6000, 6010, 6020, 6030],
    }

    api = otg.OtgApi()
    c = tcp_header_valuelist_ports_config(api, test_const)

    api.set_config(c)

    api.start_transmit()

    api.wait_for(
        fn=lambda: metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )


def tcp_header_valuelist_ports_config(api, tc):
    c = api.api.config()
    p1 = c.ports.add(name="p1", location=api.test_config.otg_ports[0])
    p2 = c.ports.add(name="p2", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[p1.name, p2.name])
    ly.speed = api.test_config.otg_speed

    f1 = c.flows.add(name="f1")
    f1.tx_rx.port.tx_name = p1.name
    f1.tx_rx.port.rx_name = p2.name
    f1.duration.fixed_packets.packets = tc["pktCount"]
    f1.rate.pps = tc["pktRate"]
    f1.size.fixed = tc["pktSize"]
    f1.metrics.enable = True

    eth, ip, tcp = f1.packet.ethernet().ipv4().tcp()

    eth.src.value = tc["txMac"]
    eth.dst.value = tc["rxMac"]

    ip.src.value = tc["txIp"]
    ip.dst.value = tc["rxIp"]

    tcp.src_port.values = tc["txTcpPortValueList"]
    tcp.dst_port.values = tc["rxTcpPortValueList"]

    log.info("Config:\n%s", c)
    return c


def metrics_ok(api, tc):
    m = api.get_flow_metrics()[0]
    ok = (
        m.transmit == m.STOPPED
        and m.frames_tx == tc["pktCount"]
        and m.frames_rx == tc["pktCount"]
    )
    return ok
