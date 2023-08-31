import ipaddress
import logging as log
import datetime
import tempfile
import time

import snappi
from pcapfile import savefile as pcap_loader

from helpers.table import table
from helpers.testconfig import config as testconfig
from helpers.plot import plot


class OtgApi(object):
    def __init__(self):
        self.test_config = testconfig.TestConfig()
        log.info("OTG Host: %s", self.test_config.otg_host)
        log.info("OTG Ports: %s", self.test_config.otg_ports)
        self.api = snappi.api(
            self.test_config.otg_host,
            verify=False,
            transport="grpc" if self.test_config.otg_grpc_transport else "http",
            version_check=True,
        )
        self.plot = plot.Plot()

    def timer(self, fn_name, since):
        elapsed = (datetime.datetime.now() - since).microseconds * 1000
        self.plot.append_duration(plot.Duration(fn_name, elapsed, since))
        log.info("Elapsed duration %s: %d ns", fn_name, elapsed)

    def wait_for(
        self, fn, fn_name="wait_for", interval_seconds=0.5, timeout_seconds=10
    ):
        start = datetime.datetime.now()
        try:
            log.info("Waiting for %s ...", fn_name)
            while True:
                if fn():
                    log.info("Done waiting for %s", fn_name)
                    return

                elapsed = datetime.datetime.now() - start
                if elapsed.seconds > timeout_seconds:
                    msg = "timeout occurred while waiting for %s" % fn_name
                    raise Exception(msg)

                time.sleep(interval_seconds)
        finally:
            self.timer(fn_name, start)

    def mac_addr_to_bytes(self, mac):
        splits = mac.split(":")
        if len(splits) != 6:
            raise Exception("Invalid MAC address: " + mac)

        return bytearray([int(x, 16) for x in splits])

    def ipv4_addr_to_bytes(self, ip):
        return self.num_to_bytes(int(ipaddress.ip_address(ip)), 4)

    def ipv6_addr_to_bytes(self, ip):
        return self.num_to_bytes(int(ipaddress.ip_address(ip)), 16)

    def num_to_bytes(self, num, size):
        return num.to_bytes(size, "big")

    def log_warn(self, response):
        if response and response.warnings:
            for w in response.warnings:
                log.warning(w)

    def log_plot(self, name):
        self.plot.analyze(name)
        log.info("plot: %s\n", self.plot.to_json())

    def new_config_from_json(self, json_str):
        start = datetime.datetime.now()
        try:
            log.info("Loading config from JSON ...")
            c = self.api.config()
            c.deserialize(json_str)
            return c
        finally:
            self.timer("new_config_from_json", start)

    def new_config_from_yaml(self, yaml_str):
        start = datetime.datetime.now()
        try:
            log.info("Loading config from YAML ...")
            c = self.api.config()
            c.deserialize(yaml_str)
            return c
        finally:
            self.timer("new_config_from_yaml", start)

    def config_to_json(self, c):
        start = datetime.datetime.now()
        try:
            log.info("Serializing config to JSON ...")
            return c.serialize(encoding=c.JSON)
        finally:
            self.timer("config_to_json", start)

    def config_to_yaml(self, c):
        start = datetime.datetime.now()
        try:
            log.info("Serializing config to YAML ...")
            return c.serialize(encoding=c.YAML)
        finally:
            self.timer("config_to_yaml", start)

    def set_config(self, c):
        start = datetime.datetime.now()
        try:
            log.info("Setting config ...")
            self.log_warn(self.api.set_config(c))
        finally:
            self.timer("set_config", start)

    def cleanup_config(self):
        log.info("Cleaning up config ...")
        self.set_config(self.api.new_config())

    def start_protocols(self):
        start = datetime.datetime.now()
        try:
            log.info("Starting protocols ...")
            cs = self.api.control_state()
            cs.protocol.all.state = cs.protocol.all.START
            self.log_warn(self.api.set_control_state(cs))
        finally:
            self.timer("start_protocols", start)

    def stop_protocols(self):
        start = datetime.datetime.now()
        try:
            log.info("Stopping protocols ...")
            cs = self.api.control_state()
            cs.protocol.all.state = cs.protocol.all.STOP
            self.log_warn(self.api.set_control_state(cs))
        finally:
            self.timer("stop_protocols", start)

    def start_transmit(self):
        start = datetime.datetime.now()
        try:
            log.info("Starting transmit ...")
            cs = self.api.control_state()
            cs.traffic.flow_transmit.state = cs.traffic.flow_transmit.START
            self.log_warn(self.api.set_control_state(cs))
        finally:
            self.timer("start_transmit", start)

    def stop_transmit(self):
        start = datetime.datetime.now()
        try:
            log.info("Stopping transmit ...")
            cs = self.api.control_state()
            cs.traffic.flow_transmit.state = cs.traffic.flow_transmit.STOP
            self.log_warn(self.api.set_control_state(cs))
        finally:
            self.timer("stop_transmit", start)

    def start_capture(self):
        if not self.test_config.otg_capture_check:
            log.info("Skipped start_capture")
            return
        start = datetime.datetime.now()
        try:
            log.info("Starting capture ...")
            cs = self.api.control_state()
            cs.port.capture.state = cs.port.capture.START
            self.log_warn(self.api.set_control_state(cs))
        finally:
            self.timer("start_capture", start)

    def stop_capture(self):
        if not self.test_config.otg_capture_check:
            log.info("Skipped stop_capture")
            return
        start = datetime.datetime.now()
        try:
            log.info("Stopping capture ...")
            cs = self.api.control_state()
            cs.port.capture.state = cs.port.capture.STOP
            self.log_warn(self.api.set_control_state(cs))
        finally:
            self.timer("stop_capture", start)

    def get_flow_metrics(self):
        start = datetime.datetime.now()
        try:
            log.info("Getting flow metrics ...")
            req = self.api.metrics_request()
            req.flow.flow_names = []

            metrics = self.api.get_metrics(req).flow_metrics

            tb = table.Table(
                "Flow Metrics",
                [
                    "Name",
                    "State",
                    "Frames Tx",
                    "Frames Rx",
                    "FPS Tx",
                    "FPS Rx",
                    "Bytes Tx",
                    "Bytes Rx",
                ],
            )

            for m in metrics:
                tb.append_row(
                    [
                        m.name,
                        m.transmit,
                        m.frames_tx,
                        m.frames_rx,
                        m.frames_tx_rate,
                        m.frames_rx_rate,
                        m.bytes_tx,
                        m.bytes_rx,
                    ]
                )

            log.info(tb)
            return metrics
        finally:
            self.timer("get_flow_metrics", start)

    def get_bgpv4_metrics(self):
        start = datetime.datetime.now()
        try:
            log.info("Getting bgpv4 metrics ...")
            req = self.api.metrics_request()
            req.bgpv4.peer_names = []

            metrics = self.api.get_metrics(req).bgpv4_metrics

            tb = table.Table(
                "BGPv4 Metrics",
                [
                    "Name",
                    "State",
                    "Routes Adv.",
                    "Routes Rec.",
                ],
            )

            for m in metrics:
                tb.append_row(
                    [
                        m.name,
                        m.session_state,
                        m.routes_advertised,
                        m.routes_received,
                    ]
                )

            log.info(tb)
            return metrics
        finally:
            self.timer("get_bgpv4_metrics", start)

    def get_isis_metrics(self):
        start = datetime.datetime.now()
        try:
            log.info("Getting isis metrics ...")
            req = self.api.metrics_request()
            req.isis.router_names = []

            metrics = self.api.get_metrics(req).isis_metrics

            tb = table.Table(
                "ISIS Metrics",
                [
                    "Name",
                    "L1 Sessions Up",
                    "L2 Sessions UP",
                    "L1 Database Size",
                    "L2 Database Size",
                ],
                20,
            )

            for m in metrics:
                tb.append_row(
                    [
                        m.name,
                        m.l1_sessions_up,
                        m.l2_sessions_up,
                        m.l1_database_size,
                        m.l2_database_size,
                    ]
                )

            log.info(tb)
            return metrics
        finally:
            self.timer("get_isis_metrics", start)

    def get_ipv4_neighbors(self):
        start = datetime.datetime.now()
        try:
            log.info("Getting IPv4 Neighbors ...")
            req = self.api.states_request()
            req.ipv4_neighbors.ethernet_names = []
            neighbors = self.api.get_states(req).ipv4_neighbors

            tb = table.Table(
                "IPv4 Neighbors",
                [
                    "Ethernet Name",
                    "IPv4 Address",
                    "Link Layer Address",
                ],
                20,
            )

            for n in neighbors:
                tb.append_row(
                    [
                        n.ethernet_name,
                        n.ipv4_address,
                        "" if n.link_layer_address is None else n.link_layer_address,
                    ]
                )

            log.info(tb)
            return neighbors

        finally:
            self.timer("get_ipv4_neighbors", start)

    def get_bgp_prefixes(self):
        start = datetime.datetime.now()
        try:
            log.info("Getting BGP prefixes ...")
            req = self.api.states_request()
            req.bgp_prefixes.bgp_peer_names = []
            bgp_prefixes = self.api.get_states(req).bgp_prefixes

            tb = table.Table(
                "BGP Prefixes",
                [
                    "Name",
                    "IPv4 Address",
                    "IPv4 Next Hop",
                    "IPv6 Address",
                    "IPv6 Next Hop",
                ],
                20,
            )

            for b in bgp_prefixes:
                for p in b.ipv4_unicast_prefixes:
                    tb.append_row(
                        [
                            b.bgp_peer_name,
                            "{}/{}".format(p.ipv4_address, p.prefix_length),
                            p.ipv4_next_hop,
                            "",
                            "" if p.ipv6_next_hop is None else p.ipv6_next_hop,
                        ]
                    )
                for p in b.ipv6_unicast_prefixes:
                    tb.append_row(
                        [
                            b.bgp_peer_name,
                            "",
                            "" if p.ipv4_next_hop is None else p.ipv4_next_hop,
                            "{}/{}".format(p.ipv6_address, p.prefix_length),
                            p.ipv6_next_hop,
                        ]
                    )

            log.info(tb)
            return bgp_prefixes

        finally:
            self.timer("get_bgp_prefixes", start)

    def get_isis_lsps(self):
        start = datetime.datetime.now()
        try:
            log.info("Getting ISIS LSPs ...")
            req = self.api.states_request()
            req.isis_lsps.isis_router_names = []
            isis_lsps = self.api.get_states(req).isis_lsps

            tb = table.Table(
                "ISIS LSPs",
                [
                    "Name",
                    "LSP ID",
                    "PDU Type",
                    "IS Type",
                ],
                30,
            )

            for n in isis_lsps:
                for l in n.lsps:
                    tb.append_row(
                        [
                            n.isis_router_name,
                            l.lsp_id,
                            l.pdu_type,
                            l.is_type,
                        ]
                    )

            log.info(tb)
            return isis_lsps

        finally:
            self.timer("get_isis_lsps", start)

    def get_capture(self, port_name):
        if not self.test_config.otg_capture_check:
            log.info("Skipped get_capture")
            return None

        start = datetime.datetime.now()
        try:
            log.info("Getting capture for port %s ...", port_name)
            req = self.api.capture_request()
            req.port_name = port_name

            b = self.api.get_capture(req)
            return CapturedPackets(b)
        finally:
            self.timer("get_capture", start)


class CapturedPackets(object):
    def __init__(self, pcap_bytes):
        self.packets = []

        tmp = tempfile.TemporaryFile()
        try:
            tmp.write(pcap_bytes.read())
            tmp.seek(0)
            pcap = pcap_loader.load_savefile(tmp)

            for i, p in enumerate(pcap.packets):
                # TODO: add timestamps and lengths
                self.packets.append(CapturedPacket(i, p.raw()))
        finally:
            tmp.close()

    def validate_field(self, name, sequence, start_offset, field):
        if sequence >= len(self.packets):
            raise Exception(
                "%s: sequence %d >= len(packets) %d"
                % (name, sequence, len(self.packets))
            )

        p = self.packets[sequence]
        if start_offset < 0 or start_offset >= len(p.data):
            raise Exception(
                "%s: start_offset %d not in range [0, %d); data: %s"
                % (name, sequence, len(p.data), p.data)
            )

        end_offset = start_offset + len(field) - 1
        if end_offset < start_offset:
            raise Exception(
                "%s: start_offset %d > end_offset %d; field: %s"
                % (name, start_offset, end_offset, field)
            )

        if end_offset >= len(p.data):
            raise Exception(
                "%s: end_offset %d not in range [0, %d); field: %s data: %s"
                % (name, end_offset, len(p.data), field, p.data)
            )

        if field != p.data[start_offset : end_offset + 1]:
            raise Exception(
                "%s: field %s != actual_field %s; sequence: %d, start_offset: %d, data: %s"
                % (
                    name,
                    field,
                    p.data[start_offset : end_offset + 1],
                    sequence,
                    start_offset,
                    p.data,
                )
            )

    def validate_size(self, sequence, size):
        if len(self.packets[sequence].data) != size:
            raise Exception(
                "exp_size %d != act_size %d" % (size, len(self.packets[sequence].data))
            )

    def has_field(self, name, sequence, start_offset, field):
        try:
            self.validate_field(name, sequence, start_offset, field)
            return True
        except Exception as e:
            log.warning(e)
            return False


class CapturedPacket(object):
    def __init__(self, sequence, data):
        self.sequence = sequence
        self.data = data
