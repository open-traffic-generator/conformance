import logging as log
import pytest
from helpers.otg import otg

@pytest.mark.all
@pytest.mark.cpdp

def test_lldp_neighbors():
    test_const = {
        "txMac":       "00:00:01:01:01:01",
		"rxMac":       "00:00:01:01:01:02",
		"holdTime":    120,
		"advInterval": 5,
		"pduCount":    2,
    }

    api = otg.OtgApi()
    c = lldp_neighbor_config(api, test_const)

    api.set_config(c)

    api.start_protocols()

    api.wait_for(
        fn=lambda: lldp_neigbhors_metrics_ok(api, test_const),
        fn_name="wait_for_lldp_metrics",
        timeout_seconds=30,
    )

    api.wait_for(
        fn=lambda: lldp_neighbors_ok(api, test_const),
        fn_name="wait_for_lldp_neighbors",
        timeout_seconds=30,
    )
    

def lldp_neighbor_config(api, tc):
    c = api.api.config()

    ptx = c.ports.add(name="ptx", location=api.test_config.otg_ports[0])
    prx = c.ports.add(name="prx", location=api.test_config.otg_ports[1])

    ly = c.layer1.add(name="ly", port_names=[ptx.name, prx.name])
    ly.speed = api.test_config.otg_speed

    lldp_Tx = c.lldp.add(name="lldp_tx")
    lldp_Rx = c.lldp.add(name="lldp_rx")

    lldp_Tx.hold_time = tc["holdTime"]
    lldp_Tx.advertisement_interval = tc["advInterval"]
    lldp_Tx.connection.port_name = ptx.name
    lldp_Tx.chassis_id.mac_address_subtype.value = tc["txMac"]

    lldp_Rx.hold_time = tc["holdTime"]
    lldp_Rx.advertisement_interval = tc["advInterval"]
    lldp_Rx.connection.port_name = prx.name
    lldp_Rx.chassis_id.mac_address_subtype.value = tc["rxMac"]

    log.info("Config:\n%s", c)
    return c

def lldp_neigbhors_metrics_ok(api, tc):
    for m in api.get_lldp_metrics():
        if (
            m.frames_tx < tc["pduCount"]
            or m.frames_rx < tc["pduCount"]
        ):
            return False
    return True

def lldp_neighbors_ok(api, tc):
    count = 0
    # Print LLDP neighbors TLV chassis_id, chassis_type, Port ID, Port ID Type, TTL, System name and System description
    # Validate chassis_id, Chassis_id_type and TTL
    for n in api.get_lldp_neighbors():
        for i in ["txMac", "rxMac"]:
            if n.chassis_id_type == "mac_address" and n.chassis_id == tc[i] and n.ttl == tc["holdTime"]:
                count += 1

    return count == 2
