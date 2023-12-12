import time
import snappi
import logging as log
import datetime


def test_quickstart_bgp():
    # Create a new API handle to make API calls against OTG
    # with HTTP as default transport protocol
    # api = snappi.api(location="https://localhost:8443")
    # api = snappi.api(location="https://nanorbit0.lbj.is.keysight.com:50087")

    apis = snappi.api(
        "https://nanorbit0.lbj.is.keysight.com:50087",
        verify=False,
        transport="http",
        version_check=False,
    )

    test_const = {
        "pktRate": 50,
        "pktCount": 100,
        "pktSize": 128,
        "txMac": "00:00:01:01:01:01",
        "txIp": "172.30.1.1",
        "txGateway": "172.30.1.0",
        "txPrefix": 31,
        "txAs": 65001,
        "rxMac": "00:00:01:01:01:02",
        "rxIp": "172.30.1.3",
        "rxGateway": "172.30.1.2",
        "rxPrefix": 31,
        "rxAs": 65002,
        "txRouteCount": 10,
        "rxRouteCount": 10,
        "txAdvRouteV4": "100.1.1.1",
        "rxAdvRouteV4": "200.1.1.1",
        "txVlan": 100,
        "rxVlan": 101,
    }
    # Create a new traffic configuration that will be set on OTG
    cfg = bgp_route_prefix_config(apis, test_const)

    # Optionally, print JSON representation of config
    # print("\nCONFIGURATION", cfg.serialize(encoding=cfg.JSON), sep="\n")

    # Push traffic configuration constructed so far to OTG
    apis.set_config(cfg)

    # start protocols
    cs = apis.control_state()
    cs.protocol.all.state = cs.protocol.all.START
    apis.set_control_state(cs)
    time.sleep(5)

    # Fetch BGP metrics
    def bgp_metrics_ok():
        # Fetch metrics for bgpv4
        req = apis.metrics_request()
        req.bgpv4.peer_names = []
        metrics = apis.get_metrics(req).bgpv4_metrics
        for m in metrics:
            print("BGPv4 METRICS", m, sep="\n")
            if (
                m.session_state == m.DOWN
                and m.routes_advertised != test_const["txRouteCount"]
                and m.routes_received != test_const["rxRouteCount"]
            ):
                return False
        return True

    # Keep polling until either expectation is met or deadline exceeds
    deadline = time.time() + 60
    while not bgp_metrics_ok():
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


def bgp_route_prefix_config(apis, tc):
    c = apis.config()
    ptx = c.ports.add(
        name="ptx",
        location="uhd://tf2-qa6.lbj.is.keysight.com:7531;1+nanorbit0.lbj.is.keysight.com:50071",
    )
    prx = c.ports.add(
        name="prx",
        location="uhd://tf2-qa6.lbj.is.keysight.com:7531;2+nanorbit0.lbj.is.keysight.com:50072",
    )

    c.layer1.add(
        name="port_settings", port_names=[ptx.name, prx.name], speed="speed_100_gbps"
    )

    # UHD port 3 configuration
    # adding devices
    dtx = c.devices.add(name="dtx")
    drx = c.devices.add(name="drx")

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
    dtx_ip.set(address=tc["txIp"], gateway=tc["txGateway"], prefix=tc["txPrefix"])

    # bgp configuration
    dtx.bgp.router_id = tc["txIp"]
    dtx_bgpv4 = dtx.bgp.ipv4_interfaces.add(ipv4_name=dtx_ip.name)

    dtx_bgpv4_peer = dtx_bgpv4.peers.add(name="dtx_bgpv4_peer")
    dtx_bgpv4_peer.set(
        as_number=tc["txAs"], as_type=dtx_bgpv4_peer.IBGP, peer_address=tc["txGateway"]
    )
    dtx_bgpv4_peer.learned_information_filter.set(
        unicast_ipv4_prefix=True, unicast_ipv6_prefix=True
    )

    dtx_bgpv4_peer_rrv4 = dtx_bgpv4_peer.v4_routes.add(name="dtx_bgpv4_peer_rrv4")

    dtx_bgpv4_peer_rrv4.addresses.add(
        address=tc["txAdvRouteV4"], prefix=32, count=tc["txRouteCount"], step=1
    )

    dtx_bgpv4_peer_rrv4.advanced.set(
        multi_exit_discriminator=50, origin=dtx_bgpv4_peer_rrv4.advanced.IGP
    )

    # UHD port 4 configuration
    # adding Ethernet
    drx_eth = drx.ethernets.add(name="drx_eth")
    drx_eth.connection.port_name = prx.name
    drx_eth.mac = tc["rxMac"]
    drx_eth.mtu = 1500
    # adding vlan
    drx_vlan = dtx_eth.vlans.add(name="rxVlan")
    drx_vlan.set(id=tc["rxVlan"])
    # adding ipv4
    drx_ip = drx_eth.ipv4_addresses.add(name="drx_ip")
    drx_ip.set(address=tc["rxIp"], gateway=tc["rxGateway"], prefix=tc["rxPrefix"])
    # adding bgp
    drx.bgp.router_id = tc["rxIp"]

    drx_bgpv4 = drx.bgp.ipv4_interfaces.add()
    drx_bgpv4.ipv4_name = drx_ip.name

    drx_bgpv4_peer = drx_bgpv4.peers.add(name="drx_bgpv4_peer")
    drx_bgpv4_peer.set(
        as_number=tc["rxAs"], as_type=drx_bgpv4_peer.EBGP, peer_address=tc["rxGateway"]
    )
    drx_bgpv4_peer.learned_information_filter.set(
        unicast_ipv4_prefix=True, unicast_ipv6_prefix=True
    )

    drx_bgpv4_peer_rrv4 = drx_bgpv4_peer.v4_routes.add(name="drx_bgpv4_peer_rrv4")
    drx_bgpv4_peer_rrv4.addresses.add(
        address=tc["rxAdvRouteV4"], prefix=32, count=tc["rxRouteCount"], step=1
    )

    drx_bgpv4_peer_rrv4.advanced.set(
        multi_exit_discriminator=50, origin=drx_bgpv4_peer_rrv4.advanced.EGP
    )

    drx_bgpv4_peer_rrv4.as_path.as_set_mode = drx_bgpv4_peer_rrv4.as_path.INCLUDE_AS_SET

    drx_bgpv4_peer_rrv4_seg = drx_bgpv4_peer_rrv4.as_path.segments.add()
    drx_bgpv4_peer_rrv4_seg.set(
        as_numbers=[65003, 65004], type=drx_bgpv4_peer_rrv4_seg.AS_SEQ
    )

    # flow configuration
    # flow1 --> from iBGP to eBGP
    # flow2 --> from eBGP to iBGP
    for i in range(0, 2):
        f = c.flows.add()
        f.duration.fixed_packets.packets = tc["pktCount"]
        f.rate.pps = tc["pktRate"]
        f.size.fixed = tc["pktSize"]
        f.metrics.enable = True

    ftx_v4 = c.flows[0]
    ftx_v4.name = "iBGP to eBGP"
    ftx_v4.tx_rx.device.set(
        tx_names=[dtx_bgpv4_peer_rrv4.name], rx_names=[drx_bgpv4_peer_rrv4.name]
    )

    ftx_v4_eth, ftx_v4_ip = ftx_v4.packet.ethernet().ipv4()
    ftx_v4_vlan = ftx_v4.packet.ethernet().vlan()[-1]
    ftx_v4_eth.src.value = dtx_eth.mac
    ftx_v4_eth.dst.value = drx_eth.mac
    ftx_v4_vlan.id.value = tc["txVlan"]
    ftx_v4_vlan.tpid.value = 33024
    ftx_v4_ip.src.value = tc["txAdvRouteV4"]
    ftx_v4_ip.dst.value = tc["rxAdvRouteV4"]

    frx_v4 = c.flows[1]
    frx_v4.name = "eBGP to iBGP"
    frx_v4.tx_rx.device.set(
        tx_names=[drx_bgpv4_peer_rrv4.name], rx_names=[dtx_bgpv4_peer_rrv4.name]
    )

    frx_v4_eth, frx_v4_ip = frx_v4.packet.ethernet().ipv4()
    frx_v4_vlan = frx_v4.packet.ethernet().vlan()[-1]
    frx_v4_eth.src.value = drx_eth.mac
    frx_v4_eth.dst.value = dtx_eth.mac
    frx_v4_vlan.id.value = tc["rxVlan"]
    frx_v4_vlan.tpid.value = 33024
    frx_v4_ip.src.value = tc["rxAdvRouteV4"]
    frx_v4_ip.dst.value = tc["txAdvRouteV4"]

    log.info("Config:\n%s", c)
    return c
