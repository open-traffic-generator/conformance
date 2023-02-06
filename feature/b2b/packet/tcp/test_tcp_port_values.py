import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.all
@pytest.mark.dp
def test_tcp_port_values():
    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "rxMac": "00:00:01:01:01:02",
        "txIp": "1.1.1.1",
        "rxIp": "1.1.1.2",
        "txTcpPortValues": [5000, 5010, 5020, 5030],
        "rxTcpPortValues": [6000, 6010, 6020, 6030],
    }
    api = otg.OtgApi()
    c = tcp_port_values_config(api, test_const)

    api.set_config(c)

    api.start_capture()
    api.start_transmit()

    api.wait_for(
        fn=lambda: metrics_ok(api, test_const), fn_name="wait_for_flow_metrics"
    )

    api.stop_capture()

    capture_ok(api, c, test_const)


def tcp_port_values_config(api, tc):
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

    eth, ip, tcp = f1.packet.ethernet().ipv4().tcp()

    eth.src.value = tc["txMac"]
    eth.dst.value = tc["rxMac"]

    ip.src.value = tc["txIp"]
    ip.dst.value = tc["rxIp"]

    tcp.src_port.values = tc["txTcpPortValues"]
    tcp.dst_port.values = tc["rxTcpPortValues"]

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

    ignored_count = 0
    captured_packets = api.get_capture(c.ports[1].name)

    for i, p in enumerate(captured_packets.packets):
        # ignore unexpected packets based on ethernet src MAC
        if not captured_packets.has_field(
            "ethernet src", i, 6, api.mac_addr_to_bytes(tc["txMac"])
        ):
            ignored_count += 1
            continue

        # packet size
        captured_packets.validate_size(i, tc["pktSize"])

        # ethernet header
        captured_packets.validate_field(
            "ethernet dst", i, 0, api.mac_addr_to_bytes(tc["rxMac"])
        )
        captured_packets.validate_field(
            "ethernet type", i, 12, api.num_to_bytes(2048, 2)
        )
        # ipv4 header
        captured_packets.validate_field(
            "ipv4 total length", i, 16, api.num_to_bytes(tc["pktSize"] - 14 - 4, 2)
        )
        captured_packets.validate_field("ipv4 protocol", i, 23, api.num_to_bytes(6, 1))
        captured_packets.validate_field(
            "ipv4 src", i, 26, api.ipv4_addr_to_bytes(tc["txIp"])
        )
        captured_packets.validate_field(
            "ipv4 dst", i, 30, api.ipv4_addr_to_bytes(tc["rxIp"])
        )
        # tcp header
        j = i - ignored_count
        captured_packets.validate_field(
            "tcp src",
            i,
            34,
            api.num_to_bytes(
                tc["txTcpPortValues"][j % len(tc["txTcpPortValues"])],
                2,
            ),
        )
        captured_packets.validate_field(
            "tcp dst",
            i,
            36,
            api.num_to_bytes(
                tc["rxTcpPortValues"][j % len(tc["rxTcpPortValues"])],
                2,
            ),
        )
        captured_packets.validate_field(
            "tcp data offset", i, 46, api.num_to_bytes(80, 1)
        )

    exp_count = tc["pktCount"]
    act_count = len(captured_packets.packets) - ignored_count
    if exp_count != act_count:
        raise Exception("exp_count %d != act_count %d" % (exp_count, act_count))
