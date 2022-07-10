import logging as log
import datetime
import time

import snappi

from helpers.table import table
from helpers.testconfig import config as testconfig
from helpers.plot import plot


class OtgApi(object):
    def __init__(self):
        self.test_config = testconfig.TestConfig()
        log.info("OTG Host: %s", self.test_config.otg_host)
        log.info("OTG Ports: %s", self.test_config.otg_ports)
        self.api = snappi.api(self.test_config.otg_host, verify=False)
        self.plot = plot.Plot()

    def timer(self, fn_name, since):
        elapsed = (datetime.datetime.now() - since).microseconds / 1000
        self.plot.append_duration(plot.Duration(fn_name, elapsed, since))
        log.info("Elapsed duration %s: %d ms", fn_name, elapsed)

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

    def log_warn(self, response):
        if response and response.warnings:
            for w in response.warnings:
                log.warning(w)

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
            ps = self.api.protocol_state()
            ps.state = ps.START
            self.log_warn(self.api.set_protocol_state(ps))
        finally:
            self.timer("start_protocols", start)

    def stop_protocols(self):
        start = datetime.datetime.now()
        try:
            log.info("Stopping protocols ...")
            ps = self.api.protocol_state()
            ps.state = ps.STOP
            self.log_warn(self.api.set_protocol_state(ps))
        finally:
            self.timer("stop_protocols", start)

    def start_transmit(self):
        start = datetime.datetime.now()
        try:
            log.info("Starting transmit ...")
            ts = self.api.transmit_state()
            ts.state = ts.START
            self.log_warn(self.api.set_transmit_state(ts))
        finally:
            self.timer("start_transmit", start)

    def stop_transmit(self):
        start = datetime.datetime.now()
        try:
            log.info("Stopping transmit ...")
            ts = self.api.transmit_state()
            ts.state = ts.STOP
            self.log_warn(self.api.set_transmit_state(ts))
        finally:
            self.timer("stop_transmit", start)

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
