import logging as log
import pytest
from helpers.otg import otg


@pytest.mark.all
def test_config_serdes_perf():
    test_const = {
        "flowCounts": [1, 2, 16, 64, 128, 256, 1024],
        "flowCount": 1,
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

    dist_tables = []

    for flow_count in test_const["flowCounts"]:
        test_const["flowCount"] = flow_count
        test_case = "config_serdes_%d_flows" % (2 * flow_count)

        api = otg.OtgApi()
        c = config_serdes_perf_config(api, test_const)

        c_json = c.serialize(encoding=c.JSON)
        c_yaml = c.serialize(encoding=c.YAML)

        log.info("TEST CASE: %s", test_case)
        for i in range(1, api.test_config.otg_iterations + 1):
            log.info("ITERATION: %d\n\n", i)

            api.config_to_json(api.new_config_from_json(c_json))
            api.config_to_yaml(api.new_config_from_yaml(c_yaml))

            api.plot.append_zero()

        api.log_plot(test_case)
        dist_tables.append(api.plot.to_table())

    for d in dist_tables:
        log.info(d)


def config_serdes_perf_config(api, tc):
    c = api.api.config()
    p1 = c.ports.add(name="p1", location=api.test_config.otg_ports[0])
    p2 = c.ports.add(name="p2", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[p1.name, p2.name])
    ly.speed = api.test_config.otg_speed

    if api.test_config.otg_capture_check:
        ca = c.captures.add(name="ca", port_names=[p1.name, p2.name])
        ca.format = ca.PCAP

    for i in range(1, tc["flowCount"] + 1):
        f = c.flows.add(name="f%s-%d" % (p1.name, i))
        f.tx_rx.port.tx_name = p1.name
        f.tx_rx.port.rx_name = p2.name
        f.duration.fixed_packets.packets = tc["pktCount"]
        f.rate.pps = tc["pktRate"]
        f.size.fixed = tc["pktSize"]
        f.metrics.enable = True

        eth, ip, udp = f.packet.ethernet().ipv4().udp()

        eth.src.value = tc["txMac"]
        eth.dst.value = tc["rxMac"]

        ip.src.value = tc["txIp"]
        ip.dst.value = tc["rxIp"]

        udp.src_port.value = tc["txUdpPort"]
        udp.dst_port.value = tc["rxUdpPort"]

    for i in range(1, tc["flowCount"] + 1):
        f = c.flows.add(name="f%s-%d" % (p2.name, i))
        f.tx_rx.port.tx_name = p2.name
        f.tx_rx.port.rx_name = p1.name
        f.duration.fixed_packets.packets = tc["pktCount"]
        f.rate.pps = tc["pktRate"]
        f.size.fixed = tc["pktSize"]
        f.metrics.enable = True

        eth, ip, udp = f.packet.ethernet().ipv4().udp()

        eth.src.value = tc["rxMac"]
        eth.dst.value = tc["txMac"]

        ip.src.value = tc["rxIp"]
        ip.dst.value = tc["txIp"]

        udp.src_port.value = tc["rxUdpPort"]
        udp.dst_port.value = tc["txUdpPort"]

    log.info("Config:\n%s", c)
    return c
