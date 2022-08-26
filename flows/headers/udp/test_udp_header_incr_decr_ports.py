import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.free
@pytest.mark.b2b
def test_udp_header_incr_decr_ports():
    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "rxMac": "00:00:01:01:01:02",
        "txIp": "1.1.1.1",
        "rxIp": "1.1.1.2",
        "txUdpPortStart": 5000,
        "txUdpPortStep": 2,
        "txUdpPortCount": 10,
        "rxUdpPortStart": 6000,
        "rxUdpPortStep": 2,
        "rxUdpPortCount": 10,
    }

    api = otg.OtgApi()
    c = udp_header_incr_decr_ports_config(api, test_const)

    api.set_config(c)

    api.start_transmit()

    api.wait_for(
        fn=lambda: metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )


def udp_header_incr_decr_ports_config(api, tc):
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

    eth, ip, udp = f1.packet.ethernet().ipv4().udp()

    eth.src.value = tc["txMac"]
    eth.dst.value = tc["rxMac"]

    ip.src.value = tc["txIp"]
    ip.dst.value = tc["rxIp"]

    udp.src_port.increment.start = tc["txUdpPortStart"]
    udp.src_port.increment.step = tc["txUdpPortStep"]
    udp.src_port.increment.count = tc["txUdpPortCount"]

    udp.dst_port.decrement.start = tc["rxUdpPortStart"]
    udp.dst_port.decrement.step = tc["rxUdpPortStep"]
    udp.dst_port.decrement.count = tc["rxUdpPortCount"]

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
