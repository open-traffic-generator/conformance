# Open Traffic Generator Conformance Tests

This repository hosts equivalent Go and Python tests written using [snappi](https://github.com/open-traffic-generator/snappi) against a back-to-back connected [Ixia-C](https://github.com/open-traffic-generator/ixia-c) topology.


### Prerequisites

- Recommended OS is Ubuntu LTS release.
- At least 2 CPU cores
- At least 6GB RAM
- At least 10GB Free Hard Disk Space
- Go 1.17+ or Python 3.6+ (with pip)
- Docker Engine (Community Edition)

### Usage:

1. Clone this repository

    ```sh
    git clone https://github.com/open-traffic-generator/conformance.git && cd conformance
    ```

2. Deploy topology

    ```sh
    # use free distribution of ixia-c
    ./do.sh topo new
    # use licensed distribution of ixia-c
    ./do.sh topo new lic
    ```

3. Setup and run Go tests

    ```sh
    # setup test requirements
    ./do.sh pregotest
    # run all tests against free distribution of ixia-c
    ./do.sh gotest -tags=free
    # run all tests against licensed distribution of ixia-c
    ./do.sh gotest -tags=lic
    # run single test
    ./do.sh gotest -tags=free -run=TestUdpHeader
    ```

4. Setup and run Python tests

    ```sh
    # setup test requirements
    ./do.sh prepytest
    # run all tests against free distribution of ixia-c
    ./do.sh pytest -m free
    # run all tests against licensed distribution of ixia-c
    ./do.sh pytest -m lic
    # run single test
    ./do.sh pytest -m free -k test_udp_header
    ```

5. Teardown topology

    ```sh
    # remove free distribution of ixia-c
    ./do.sh topo rm
    # remove licensed distribution of ixia-c
    ./do.sh topo rm lic
    ```

### Advanced Usage:

1. Run perf tests in Go

    ```sh
    # run all perf tests
    ./do.sh gotest -tags=perf
    # run single perf test
    ./do.sh gotest -tags=perf -run=TestUdpHeaderMeshFlow
    # run single perf test with lesser number of iterations (default=100)
    OTG_ITERATIONS=2 ./do.sh gotest -tags=perf -run=TestUdpHeaderMeshFlow
    ```
