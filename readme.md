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
    # use DP-only distribution of ixia-c
    ./do.sh topo new dp

    # use CP/DP distribution of ixia-c
    # enter GITHUB PAT instead of password when prompted for credentials
    ./do.sh topo new cpdp
    ```

    > Once deployment is done, `test-config.yaml` is automatically generated in present working directory and can be used to customize test run

3. Setup and run Go tests

    ```sh
    # setup test requirements
    ./do.sh pregotest

    # run all feature tests against DP-only distribution of ixia-c
    ./do.sh gotest -tags="dp_feature"
    # run all feature tests against CP/DP distribution of ixia-c
    ./do.sh gotest -tags="feature"

    # run single test
    ./do.sh gotest -tags="all" -run="^TestUdpHeader$"
    ```

4. Setup and run Python tests

    ```sh
    # setup test requirements
    ./do.sh prepytest

    # run all tests against DP-only distribution of ixia-c
    ./do.sh pytest -m dp_feature
    # run all tests against CP/DP distribution of ixia-c
    ./do.sh pytest -m feature

    # run single test
    ./do.sh pytest -m all -k test_udp_header
    ```

5. Teardown topology

    ```sh
    # remove DP-only distribution of ixia-c
    ./do.sh topo rm dp

    # remove CP/DP distribution of ixia-c
    ./do.sh topo rm cpdp
    ```

6. Format python code
    
    Note that if you submit any code which does not follow proper python format the CI will fail

    ```sh
    # to format python code
    ./do.sh pylint

    # to format python code for a specific path
    ./do.sh pylint features
    ```

7. Format go code

   Note that if you submit any code which does not follow proper go format the CI will fail

    ```sh
    # to format go code
    ./do.sh golint

    # to format go code for a specific path
    ./do.sh golint helpers
    ```

### Advanced Usage:

1. Run perf tests in Go

    ```sh
    # run all perf tests
    ./do.sh gotest -tags=perf
    # run single perf test
    ./do.sh gotest -tags=perf -run=TestUdpHeaderMeshFlowsPerf
    # run single perf test with lesser number of iterations (default=100)
    OTG_ITERATIONS=2 ./do.sh gotest -tags=perf -run=TestUdpHeaderMeshFlowsPerf
    ```

2. Run tests against ixia-c B2B deployed on K8S cluster using eth0 as test interface

    ```sh
    # setup K8S cluster
    ./do.sh new_k8s_cluster
    # create topology
    ./do.sh topo new k8seth0
    # setup Go tests
    ./do.sh pregotest
    # run single test
    ./do.sh gotest -tags="all" -run="^TestUdpHeaderEth0$"
    # delete topology
    ./do.sh topo rm k8seth0
    # delete K8S cluster
    ./do.sh rm_k8s_cluster
    ```

3. Run tests against KNE cluster

    ```sh
    # setup KNE cluster
    ./do.sh new_k8s_cluster kne
    # create KNE topology
    ./do.sh topo new kneb2b
    # setup Go tests
    ./do.sh pregotest
    # run single test
    ./do.sh gotest -tags="all" -run="^TestEbgpv4RouteInstall$"
    # run all Go tests
    ./do.sh gotest -tags="feature"
    # delete KNE topology
    ./do.sh topo rm kneb2b
    # delete KNE cluster
    ./do.sh rm_k8s_cluster kne
    ```
