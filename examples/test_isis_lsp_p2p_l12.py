import time
import snappi
import datetime


def test_isis_lsp_p2p_l12():

    apis = snappi.api(
        "https://localhost:8443",
        verify=False,
        transport="http",
        version_check=True,
    )

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
        "txIsisSystemId": "640000000001",
        "txIsisAreaAddress": ["490001"],
        "rxMac": "00:00:01:01:01:02",
        "rxIp": "1.1.1.2",
        "rxGateway": "1.1.1.1",
        "rxPrefix": 24,
        "rxIpv6": "1100::2",
        "rxv6Gateway": "1100::1",
        "rxv6Prefix": 64,
        "rxIsisSystemId": "650000000001",
        "rxIsisAreaAddress": ["490001"],
        "txRouteCount": 1,
        "rxRouteCount": 1,
        "txAdvRouteV4": "10.10.10.1",
        "rxAdvRouteV4": "20.20.20.1",
        "txAdvRouteV6": "::10:10:10:01",
        "rxAdvRouteV6": "::20:20:20:01",
    }

    c = isis_lsp_p2p_l12_config(apis, test_const)
    apis.set_config(c)

    cs = apis.control_state()
    cs.protocol.all.state = cs.protocol.all.START
    apis.set_control_state(cs)
    time.sleep(10)

    deadline = 60
    session_up = 0
    for i in range(0, deadline):
        req = apis.metrics_request()
        req.isis.router_names = []
        metrics = apis.get_metrics(req).isis_metrics
        print(metrics)
        for m in metrics:
            print("ISIS METRICS", m, sep="\n")
            l1lsp = int(m.l1_lsp_received)
            l2lsp = int(m.l1_lsp_received)
            total_lsp = l1lsp + l2lsp
            if (m.l2_sessions_up == 1 and m.l1_sessions_up == 1):
                session_up = m.l2_sessions_up + m.l1_sessions_up
                if l1lsp > 0 and l2lsp > 0 and total_lsp == 8 :
                    break
                # end if
            # end if
        # end for
        if session_up == 2 and total_lsp >= 8:
            break
        time.sleep(1)
    else:
        return False
    # end for

    print("Sessions up Starting Traffic")
    cs = apis.control_state()

    cs.traffic.flow_transmit.state = cs.traffic.flow_transmit.START
    apis.set_control_state(cs)
    time.sleep(5)
    cs.traffic.flow_transmit.state = cs.traffic.flow_transmit.STOP

    traffic_stats = flow_metrics_ok(apis, test_const)

    cs.protocol.all.state = cs.protocol.all.STOP
    apis.set_control_state(cs)


def isis_lsp_p2p_l12_config(apis, tc):
    c = apis.config()
    #ptx = c.ports.add(name="ptx", location=apis.test_config.otg_ports[0])
    #prx = c.ports.add(name="prx", location=apis.test_config.otg_ports[1])
    ptx = c.ports.add(name="ptx",
        location="uhd://10.36.70.24:7531;5+10.36.70.83:50075"
    )
    prx = c.ports.add(name="prx", 
        location="uhd://10.36.70.24:7531;5+10.36.70.83:50076"
    )

    ly = c.layer1.add(name="ly", port_names=[ptx.name, prx.name])
    #ly.speed = apis.test_config.otg_speed

    # transmitter
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

    dtx_ipv6 = dtx_eth.ipv6_addresses.add(name="dtxv6_ip")
    dtx_ipv6.address = tc["txIpv6"]
    dtx_ipv6.gateway = tc["txv6Gateway"]
    dtx_ipv6.prefix = tc["txv6Prefix"]

    dtx.isis.system_id = tc["txIsisSystemId"]
    dtx.isis.name = "dtx_isis"

    dtx.isis.advanced.area_addresses = tc["txIsisAreaAddress"]
    dtx.isis.advanced.lsp_refresh_rate = 900
    dtx.isis.advanced.enable_attached_bit = False

    dtx.isis.basic.ipv4_te_router_id = tc["txIp"]
    dtx.isis.basic.hostname = dtx.isis.name
    dtx.isis.basic.learned_lsp_filter = True

    dtx_isis_int = dtx.isis.interfaces.add()
    dtx_isis_int.eth_name = dtx_eth.name
    dtx_isis_int.name = "dtx_isis_int"
    dtx_isis_int.network_type = dtx_isis_int.POINT_TO_POINT
    dtx_isis_int.level_type = dtx_isis_int.LEVEL_1_2

    dtx_isis_int.l2_settings.dead_interval = 30
    dtx_isis_int.l2_settings.hello_interval = 10
    dtx_isis_int.l2_settings.priority = 0

    dtx_isis_int.advanced.auto_adjust_supported_protocols = True

    dtx_isis_rr4 = dtx.isis.v4_routes.add(name="dtx_isis_rr4")
    dtx_isis_rr4.link_metric = 10
    dtx_isis_rr4.addresses.add(
        address=tc["txAdvRouteV4"], prefix=32, count=tc["txRouteCount"], step=1
    )

    dtx_isis_rrv6 = dtx.isis.v6_routes.add(name="dtx_isis_rr6")
    dtx_isis_rrv6.addresses.add(
        address=tc["txAdvRouteV6"], prefix=32, count=tc["txRouteCount"], step=1
    )

    # receiver
    drx_eth = drx.ethernets.add(name="drx_eth")
    drx_eth.connection.port_name = prx.name
    drx_eth.mac = tc["rxMac"]
    drx_eth.mtu = 1500

    drx_ip = drx_eth.ipv4_addresses.add(name="drx_ip")
    drx_ip.address = tc["rxIp"]
    drx_ip.gateway = tc["rxGateway"]
    drx_ip.prefix = tc["rxPrefix"]

    drx_ipv6 = drx_eth.ipv6_addresses.add(name="drxv6_ip")
    drx_ipv6.address = tc["rxIpv6"]
    drx_ipv6.gateway = tc["rxv6Gateway"]
    drx_ipv6.prefix = tc["rxv6Prefix"]

    drx.isis.system_id = tc["rxIsisSystemId"]
    drx.isis.name = "rx_isis"

    drx.isis.advanced.area_addresses = tc["rxIsisAreaAddress"]
    drx.isis.advanced.lsp_refresh_rate = 900
    drx.isis.advanced.enable_attached_bit = False

    drx.isis.basic.ipv4_te_router_id = tc["rxIp"]
    drx.isis.basic.hostname = drx.isis.name
    drx.isis.basic.learned_lsp_filter = True

    drx_isis_int = drx.isis.interfaces.add()
    drx_isis_int.eth_name = drx_eth.name
    drx_isis_int.name = "drx_isis_int"
    drx_isis_int.network_type = drx_isis_int.POINT_TO_POINT
    drx_isis_int.level_type = drx_isis_int.LEVEL_1_2

    drx_isis_int.l2_settings.dead_interval = 30
    drx_isis_int.l2_settings.hello_interval = 10
    drx_isis_int.l2_settings.priority = 0

    drx_isis_int.advanced.auto_adjust_supported_protocols = True

    drx_isis_rr4 = drx.isis.v4_routes.add(name="drx_isis_rr4")
    drx_isis_rr4.link_metric = 10
    drx_isis_rr4.addresses.add(
        address=tc["rxAdvRouteV4"], prefix=32, count=tc["rxRouteCount"], step=1
    )

    drx_isis_rrv6 = drx.isis.v6_routes.add(name="drx_isis_rr6")
    drx_isis_rrv6.addresses.add(
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
    ftx_v4.tx_rx.device.tx_names = [dtx_isis_rr4.name]
    ftx_v4.tx_rx.device.rx_names = [drx_isis_rr4.name]

    ftx_v4_eth, ftx_v4_ip, ftx_v4_tcp = ftx_v4.packet.ethernet().ipv4().tcp()
    ftx_v4_eth.src.value = dtx_eth.mac
    ftx_v4_ip.src.value = tc["txAdvRouteV4"]
    ftx_v4_ip.dst.value = tc["rxAdvRouteV4"]
    ftx_v4_tcp.src_port.value = 5000
    ftx_v4_tcp.dst_port.value = 6000

    ftx_v6 = c.flows[1]
    ftx_v6.name = "ftx_v6"
    ftx_v6.tx_rx.device.tx_names = [dtx_isis_rrv6.name]
    ftx_v6.tx_rx.device.rx_names = [drx_isis_rrv6.name]

    ftx_v6_eth, ftx_v6_ip, ftx_v6_tcp = ftx_v6.packet.ethernet().ipv6().tcp()
    ftx_v6_eth.src.value = dtx_eth.mac
    ftx_v6_ip.src.value = tc["txAdvRouteV6"]
    ftx_v6_ip.dst.value = tc["rxAdvRouteV6"]
    ftx_v6_tcp.src_port.value = 5000
    ftx_v6_tcp.dst_port.value = 6000

    frx_v4 = c.flows[2]
    frx_v4.name = "frx_v4"
    frx_v4.tx_rx.device.tx_names = [drx_isis_rr4.name]
    frx_v4.tx_rx.device.rx_names = [dtx_isis_rr4.name]

    frx_v4_eth, frx_v4_ip, frx_v4_tcp = frx_v4.packet.ethernet().ipv4().tcp()
    frx_v4_eth.src.value = drx_eth.mac
    frx_v4_ip.src.value = tc["rxAdvRouteV4"]
    frx_v4_ip.dst.value = tc["txAdvRouteV4"]
    frx_v4_tcp.src_port.value = 5000
    frx_v4_tcp.dst_port.value = 6000

    frx_v6 = c.flows[3]
    frx_v6.name = "frx_v6"
    frx_v6.tx_rx.device.tx_names = [drx_isis_rrv6.name]
    frx_v6.tx_rx.device.rx_names = [dtx_isis_rrv6.name]

    frx_v6_eth, frx_v6_ip, frx_v6_tcp = frx_v6.packet.ethernet().ipv6().tcp()
    frx_v6_eth.src.value = drx_eth.mac
    frx_v6_ip.src.value = tc["rxAdvRouteV6"]
    frx_v6_ip.dst.value = tc["txAdvRouteV6"]
    frx_v6_tcp.src_port.value = 5000
    frx_v6_tcp.dst_port.value = 6000

    print("Config:\n%s", c)
    return c


def flow_metrics_ok(apis, tc):
    # Fetch metrics for configured flow
    mr = apis.metrics_request()
    mr.flow.flow_names = []
    metrics = apis.get_metrics(mr).flow_metrics
    for m in metrics:
        print("FLOW METRICS", m, sep="\n")
        if (
            m.transmit != m.STOPPED
            or m.frames_tx != tc["pktCount"]
            or m.frames_rx != tc["pktCount"]
         ):
            return False
    return True

#test_isis_lsp_p2p_l12()
