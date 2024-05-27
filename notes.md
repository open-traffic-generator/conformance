#### Build Details

| Component                     | Version       |
|-------------------------------|---------------|
| Open Traffic Generator API    | [1.5.0](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/open-traffic-generator/models/v1.5.0/artifacts/openapi.yaml)         |
| snappi                        | [1.5.0](https://pypi.org/project/snappi/1.5.0)        |
| gosnappi                      | [1.5.0](https://pkg.go.dev/github.com/open-traffic-generator/snappi/gosnappi@v1.5.0)        |
| keng-controller               | [1.5.0-1](https://github.com/orgs/open-traffic-generator/packages/container/package/keng-controller)    |
| ixia-c-traffic-engine         | [1.8.0.7](https://github.com/orgs/open-traffic-generator/packages/container/package/ixia-c-traffic-engine)       |
| keng-app-usage-reporter       | [0.0.1-52](https://github.com/orgs/open-traffic-generator/packages/container/package/keng-app-usage-reporter)      |
| ixia-c-protocol-engine        | [1.00.0.382](https://github.com/orgs/open-traffic-generator/packages/container/package/ixia-c-protocol-engine)    | 
| keng-layer23-hw-server        | [1.5.0-1](https://github.com/orgs/open-traffic-generator/packages/container/package/keng-layer23-hw-server)    |
| keng-operator                 | [0.3.28](https://github.com/orgs/open-traffic-generator/packages/container/package/keng-operator)        | 
| otg-gnmi-server               | [1.14.1](https://github.com/orgs/open-traffic-generator/packages/container/package/otg-gnmi-server)         |
| ixia-c-one                    | [1.5.0-1](https://github.com/orgs/open-traffic-generator/packages/container/package/ixia-c-one/)         |
| UHD400                        | [1.2.7](https://downloads.ixiacom.com/support/downloads_and_updates/public/UHD400/1.2/1.2.7/artifacts.tar)         |


# Bug Fix(s)
* <b><i>Ixia-C</i></b>: An issue has been detected whereby some internal certificates used in the ixia-c containerized solution has expired. <b><i>This is fixed in this patch build.</i></b>

  The most common manifestation of this is that despite proper ARP resolution, Traffic Start in tests will fail in combined protocol-engine/traffic-engine setups with `Error occurred while starting flows: [error starting tx port <portname>: unsuccessful Response: MAC address resolution failed for IP: <ip address>  ]`.

  <b><i>This issue should not affect standalone traffic-engine setups using 'port' or raw traffic flows.</i></b>

  This issue is visible for ixia-c community releases starting from `v0.1.0-3` to `v1.4.0-15`.



#### Known Issues
* <b><i>Ixia Chassis & Appliances(Novus, AresOne)</i></b>: If `keng-layer23-hw-server` version is upgraded/downgraded, the ports which will be used from this container must be rebooted once before running the tests.
* <b><i>UHD400</i></b>: Packets will not be transmitted if `flows[i].rate.pps` is less than 50.
* <b><i>UHD400</i></b>: `values` for fields in flow packet headers can be created with maximum length of 1000 values.
* <b><i>Ixia-C</i></b>: Flow Tx is incremented for flow with tx endpoints as LAG, even if no packets are sent on the wire when all active links of the LAG are down. 
* <b><i>Ixia-C</i></b>: Supported value for `flows[i].metrics.latency.mode` is `cut_through`.
* <b><i>Ixia-C</i></b>: The metric `loss` in flow metrics is currently not supported.
* <b><i>Ixia-C</i></b>: When flow transmit is started, transmission will be restarted on any existing flows already transmitting packets. 