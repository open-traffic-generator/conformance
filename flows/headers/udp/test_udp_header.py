from helpers.otg import otg


def test_udp_header():
    c = otg.api.config()
    p1 = c.ports.port(name="p1", location=otg.otg_ports[0])[-1]
    p2 = c.ports.port(name="p2", location=otg.otg_ports[1])[-1]

    ly = c.layer1.layer1(name="ly", port_names=[p1.name, p2.name])[-1]
    ly.speed = ly.SPEED_1_GBPS

    f1 = c.flows.flow(name="f1")[-1]
    f1.tx_rx.port.tx_name = p1.name
    f1.tx_rx.port.rx_name = p2.name
    f1.duration.fixed_packets.packets = 100
    f1.rate.pps = 50
    f1.size.fixed = 128
    f1.metrics.enable = True

    eth, ip, udp = f1.packet.ethernet().ipv4().udp()

    eth.src.value = "00:00:00:00:00:AA"
    eth.dst.value = "00:00:00:00:00:BB"

    ip.src.value = "1.1.1.10"
    ip.dst.value = "1.1.1.20"

    udp.src_port.value = 5000
    udp.dst_port.value = 6000

    otg.set_config(c)

    otg.start_transmit()

    otg.wait_for(fn=metrics_ok, fn_name="wait_for_flow_metrics")


def metrics_ok():
    m = otg.get_flow_metrics()[0]
    ok = m.transmit == m.STOPPED and m.frames_tx == 100 and m.frames_rx == 100
    return ok
