import logging as log
import datetime
import time

import snappi

from helpers.table import table

otg_host = "https://localhost"
otg_port1 = "localhost:5555"
otg_port2 = "localhost:5556"
api = snappi.api(otg_host, verify=False)


def timer(fn_name, since):
    elapsed = (datetime.datetime.now() - since).microseconds / 1000
    log.info("Elapsed duration %s: %d ms" % (fn_name, elapsed))


def wait_for(fn, fn_name="wait_for", interval_seconds=0.5, timeout_seconds=10):
    start = datetime.datetime.now()
    try:
        log.info("Waiting for %s ..." % fn_name)
        while True:
            done = fn()
            if done:
                log.info("Done waiting for %s" % fn_name)
                return

            elapsed = datetime.datetime.now() - start
            if elapsed.seconds > timeout_seconds:
                msg = "timeout occurred while waiting for %s" % fn_name
                raise Exception(msg)

            time.sleep(interval_seconds)
    finally:
        timer(fn_name, start)


def log_warn(response):
    if response and response.warnings:
        for w in response.warnings:
            log.warning(w)


def set_config(c):
    start = datetime.datetime.now()
    try:
        log.info("Setting config ...")
        log_warn(api.set_config(c))
    finally:
        timer("set_config", start)


def cleanup_config():
    log.info("Cleaning up config ...")
    set_config(api.new_config())


def start_protocols():
    start = datetime.datetime.now()
    try:
        log.info("Starting protocols ...")
        ps = api.protocol_state()
        ps.state = ps.START
        log_warn(api.set_protocol_state(ps))
    finally:
        timer("start_protocols", start)


def stop_protocols():
    start = datetime.datetime.now()
    try:
        log.info("Stopping protocols ...")
        ps = api.protocol_state()
        ps.state = ps.STOP
        log_warn(api.set_protocol_state(ps))
    finally:
        timer("stop_protocols", start)


def start_transmit():
    start = datetime.datetime.now()
    try:
        log.info("Starting transmit ...")
        ts = api.transmit_state()
        ts.state = ts.START
        log_warn(api.set_transmit_state(ts))
    finally:
        timer("start_transmit", start)


def stop_transmit():
    start = datetime.datetime.now()
    try:
        log.info("Stopping transmit ...")
        ts = api.transmit_state()
        ts.state = ts.STOP
        log_warn(api.set_transmit_state(ts))
    finally:
        timer("stop_transmit", start)


def get_flow_metrics():
    start = datetime.datetime.now()
    try:
        log.info("Getting flow metrics ...")
        req = api.metrics_request()
        req.flow.flow_names = []

        metrics = api.get_metrics(req).flow_metrics

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
        timer("get_flow_metrics", start)
