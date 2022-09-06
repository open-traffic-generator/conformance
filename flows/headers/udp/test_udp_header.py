import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.all
@pytest.mark.feature
@pytest.mark.b2b
@pytest.mark.free_feature
def test_udp_header():
    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "rxMac": "00:00:01:01:01:02",
        "txIp": "1.1.1.1",
        "rxIp": "1.1.1.2",
        "txUdpPort": 5000,
        "rxUdpPort": 6000,
    }

    api = otg.OtgApi()
    c = udp_header_config(api, test_const)

    api.set_config(c)

    api.start_capture()
    api.start_transmit()

    api.wait_for(
        fn=lambda: metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )

    api.stop_capture()

    capture_ok(api, c, test_const)


def udp_header_config(api, tc):
    c = api.api.config()
    p1 = c.ports.add(name="p1", location=api.test_config.otg_ports[0])
    p2 = c.ports.add(name="p2", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[p1.name, p2.name])
    ly.speed = api.test_config.otg_speed

    if api.test_config.otg_capture_check:
        ca = c.captures.add(name="ca", port_names=[p1.name, p2.name])
        ca.format = ca.PCAP

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

    udp.src_port.value = tc["txUdpPort"]
    udp.dst_port.value = tc["rxUdpPort"]

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


def capture_ok(api, c, tc):
    if not api.test_config.otg_capture_check:
        return

    captured_packets = api.get_capture(c.ports[1].name)

    for i, p in enumerate(captured_packets.packets):
        # ethernet header
        captured_packets.validate_field(
            "ethernet dst", i, 0, api.mac_addr_to_bytes(tc["rxMac"])
        )
        captured_packets.validate_field(
            "ethernet src", i, 6, api.mac_addr_to_bytes(tc["txMac"])
        )
        captured_packets.validate_field(
            "ethernet type", i, 12, api.num_to_bytes(2048, 2)
        )
        # ipv4 header
        captured_packets.validate_field(
            "ipv4 total length", i, 16, api.num_to_bytes(tc["pktSize"] - 14 - 4, 2)
        )
        captured_packets.validate_field("ipv4 protocol", i, 23, api.num_to_bytes(17, 1))
        captured_packets.validate_field(
            "ipv4 src", i, 26, api.ipv4_addr_to_bytes(tc["txIp"])
        )
        captured_packets.validate_field(
            "ipv4 dst", i, 30, api.ipv4_addr_to_bytes(tc["rxIp"])
        )
        # udp header
        captured_packets.validate_field(
            "udp src", i, 34, api.num_to_bytes(tc["txUdpPort"], 2)
        )
        captured_packets.validate_field(
            "udp dst", i, 36, api.num_to_bytes(tc["rxUdpPort"], 2)
        )
        captured_packets.validate_field(
            "udp length", i, 38, api.num_to_bytes(tc["pktSize"] - 14 - 4 - 20, 2)
        )
