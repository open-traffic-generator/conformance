import time
import snappi
import datetime


def test_isis_dut():
    apis = snappi.api(
        "https://nanorbit0.lbj.is.keysight.com:50087",
        verify=False,
        transport="http",
        version_check=True,
    )

    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "txIp": "172.30.2.1",
        "txGateway": "172.30.2.0",
        "txPrefix": 31,
        "txIpv6": "1100::172:30:2:1",
        "txv6Gateway": "1100::172:30:2:0",
        "txv6Prefix": 64,
        "txIsisSystemId": "640000000001",
        "txIsisAreaAddress": ["490001"],
        "rxMac": "00:00:01:01:01:02",
        "rxIp": "172.30.2.3",
        "rxGateway": "172.30.2.2",
        "rxPrefix": 31,
        "rxIpv6": "1100::172:30:2:3",
        "rxv6Gateway": "1100::172:30:2:2",
        "rxv6Prefix": 64,
        "rxIsisSystemId": "650000000001",
        "rxIsisAreaAddress": ["490001"],
        "txRouteCount": 10,
        "rxRouteCount": 10,
        "txAdvRouteV4": "100.1.1.1",
        "rxAdvRouteV4": "200.1.1.1",
        "txAdvRouteV6": "::100:1:1:1",
        "rxAdvRouteV6": "::200:1:1:1",
        "txVlan": 200,
        "rxVlan": 201,
    }

    # Optionally, print JSON representation of config
    # print("\nCONFIGURATION", cfg.serialize(encoding=cfg.JSON), sep="\n")

    c = isis_dut_config(apis, test_const)
    # Push traffic configuration constructed so far to OTG
    apis.set_config(c)

    # start protocols
    cs = apis.control_state()
    cs.protocol.all.state = cs.protocol.all.START
    apis.set_control_state(cs)
    time.sleep(5)

    # Fetch metrics for isis
    def isis_metrics_ok():
        # Fetch ISIS metrics
        req = apis.metrics_request()
        req.isis.router_names = []
        metrics = apis.get_metrics(req).isis_metrics
        # print(metrics)
        for m in metrics:
            if m.l2_sessions_up != 1 and m.l1_sessions_up != 1:
                return False
            # end if
        # end for
        return True

    # Keep polling until either expectation is met or deadline exceeds
    deadline = time.time() + 60
    while not isis_metrics_ok():
        if time.time() > deadline:
            raise Exception("Deadline exceeded !")
        time.sleep(0.1)

    # Start transmitting the packets from configured flow
    cs = apis.control_state()
    cs.traffic.flow_transmit.state = cs.traffic.flow_transmit.START
    apis.set_control_state(cs)
    time.sleep(5)

    def flow_metrics_ok():
        # Fetch metrics for configured flow
        mr = apis.metrics_request()
        mr.flow.flow_names = []
        metrics = apis.get_metrics(mr).flow_metrics
        for m in metrics:
            print("FLOW METRICS", m, sep="\n")
            if m.transmit != m.STOPPED and m.frames_tx != 100 and m.frames_rx != 100:
                return False
        return True

    # Keep polling until either expectation is met or deadline exceeds
    deadline = time.time() + 30
    while not flow_metrics_ok():
        if time.time() > deadline:
            raise Exception("Deadline exceeded !")
        time.sleep(0.1)

    cs.protocol.all.state = cs.protocol.all.STOP
    apis.set_control_state(cs)


def isis_dut_config(apis, tc):
    c = apis.config()

    ptx = c.ports.add(
        name="ptx",
        location="uhd://tf2-qa6.lbj.is.keysight.com:7531;5+nanorbit0.lbj.is.keysight.com:50075",
    )
    prx = c.ports.add(
        name="prx",
        location="uhd://tf2-qa6.lbj.is.keysight.com:7531;6+nanorbit0.lbj.is.keysight.com:50076",
    )

    c.layer1.add(
        name="port_settings", port_names=[ptx.name, prx.name], speed="speed_100_gbps"
    )

    # adding devices
    dtx = c.devices.add(name="dtx")
    drx = c.devices.add(name="drx")

    # UHD port3
    # ethernet configuration
    dtx_eth = dtx.ethernets.add(name="dtx_eth")
    dtx_eth.connection.port_name = ptx.name
    dtx_eth.mac = tc["txMac"]
    dtx_eth.mtu = 1500

    # vlan configuration
    dtx_vlan = dtx_eth.vlans.add(name="txVlan")
    dtx_vlan.set(id=tc["txVlan"])

    # ipv4 configuration
    dtx_ip = dtx_eth.ipv4_addresses.add(name="dtx_ip")
    dtx_ip.address = tc["txIp"]
    dtx_ip.gateway = tc["txGateway"]
    dtx_ip.prefix = tc["txPrefix"]

    # ipv6 configuration
    dtx_ipv6 = dtx_eth.ipv6_addresses.add(name="dtxv6_ip")
    dtx_ipv6.address = tc["txIpv6"]
    dtx_ipv6.gateway = tc["txv6Gateway"]
    dtx_ipv6.prefix = tc["txv6Prefix"]

    # isis configuration
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
    dtx_isis_int.level_type = dtx_isis_int.LEVEL_1

    dtx_isis_int.advanced.auto_adjust_supported_protocols = True

    dtx_isis_rr4 = dtx.isis.v4_routes.add(name="dtx_isis_rr4")
    dtx_isis_rr4.link_metric = 11
    dtx_isis_rr4.addresses.add(
        address=tc["txAdvRouteV4"], prefix=32, count=tc["txRouteCount"], step=1
    )

    dtx_isis_rrv6 = dtx.isis.v6_routes.add(name="dtx_isis_rr6")
    dtx_isis_rrv6.addresses.add(
        address=tc["txAdvRouteV6"], prefix=32, count=tc["txRouteCount"], step=1
    )

    # UHD port4
    # ethernet configuration
    drx_eth = drx.ethernets.add(name="drx_eth")
    drx_eth.connection.port_name = prx.name
    drx_eth.mac = tc["rxMac"]
    drx_eth.mtu = 1500

    # vlan configuration
    drx_vlan = dtx_eth.vlans.add(name="rxVlan")
    drx_vlan.set(id=tc["rxVlan"])

    # ipv4 configuration
    drx_ip = drx_eth.ipv4_addresses.add(name="drx_ip")
    drx_ip.address = tc["rxIp"]
    drx_ip.gateway = tc["rxGateway"]
    drx_ip.prefix = tc["rxPrefix"]

    # ipv6 configuration
    drx_ipv6 = drx_eth.ipv6_addresses.add(name="drxv6_ip")
    drx_ipv6.address = tc["rxIpv6"]
    drx_ipv6.gateway = tc["rxv6Gateway"]
    drx_ipv6.prefix = tc["rxv6Prefix"]

    # isis configuration
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
    drx_isis_int.level_type = drx_isis_int.LEVEL_2

    drx_isis_int.l2_settings.dead_interval = 30
    drx_isis_int.l2_settings.hello_interval = 10
    drx_isis_int.l2_settings.priority = 0

    drx_isis_int.advanced.auto_adjust_supported_protocols = True

    drx_isis_rr4 = drx.isis.v4_routes.add(name="drx_isis_rr4")
    drx_isis_rr4.link_metric = 11
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

    ftx_v4_eth, ftx_v4_ip = ftx_v4.packet.ethernet().ipv4()
    ftx_v4_eth.src.value = dtx_eth.mac
    ftx_v4_vlan = ftx_v4.packet.ethernet().vlan()[-1]
    ftx_v4_vlan.id.value = tc["txVlan"]
    ftx_v4_vlan.tpid.value = 33024
    ftx_v4_ip.src.value = tc["txAdvRouteV4"]
    ftx_v4_ip.dst.value = tc["rxAdvRouteV4"]

    ftx_v6 = c.flows[1]
    ftx_v6.name = "ftx_v6"
    ftx_v6.tx_rx.device.tx_names = [dtx_isis_rrv6.name]
    ftx_v6.tx_rx.device.rx_names = [drx_isis_rrv6.name]

    ftx_v6_eth, ftx_v6_ip = ftx_v6.packet.ethernet().ipv6()
    ftx_v6_eth.src.value = dtx_eth.mac
    ftx_v6_vlan = ftx_v6.packet.ethernet().vlan()[-1]
    ftx_v6_vlan.id.value = tc["txVlan"]
    ftx_v6_vlan.tpid.value = 33024
    ftx_v6_ip.src.value = tc["txAdvRouteV6"]
    ftx_v6_ip.dst.value = tc["rxAdvRouteV6"]

    frx_v4 = c.flows[2]
    frx_v4.name = "frx_v4"
    frx_v4.tx_rx.device.tx_names = [drx_isis_rr4.name]
    frx_v4.tx_rx.device.rx_names = [dtx_isis_rr4.name]

    frx_v4_eth, frx_v4_ip = frx_v4.packet.ethernet().ipv4()
    frx_v4_eth.src.value = drx_eth.mac
    ftx_v4_vlan = frx_v4.packet.ethernet().vlan()[-1]
    ftx_v4_vlan.id.value = tc["rxVlan"]
    ftx_v4_vlan.tpid.value = 33024
    frx_v4_ip.src.value = tc["rxAdvRouteV4"]
    frx_v4_ip.dst.value = tc["txAdvRouteV4"]

    frx_v6 = c.flows[3]
    frx_v6.name = "frx_v6"
    frx_v6.tx_rx.device.tx_names = [drx_isis_rrv6.name]
    frx_v6.tx_rx.device.rx_names = [dtx_isis_rrv6.name]

    frx_v6_eth, frx_v6_ip = frx_v6.packet.ethernet().ipv6()
    frx_v6_eth.src.value = drx_eth.mac
    ftx_v6_vlan = frx_v6.packet.ethernet().vlan()[-1]
    ftx_v6_vlan.id.value = tc["rxVlan"]
    ftx_v6_vlan.tpid.value = 33024
    frx_v6_ip.src.value = tc["rxAdvRouteV6"]
    frx_v6_ip.dst.value = tc["txAdvRouteV6"]

    print("Config:\n%s", c)
    return c
