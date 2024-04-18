# Open Traffic Generator Conformance Tests 

This repository hosts equivalent Go and Python tests written using [snappi](https://github.com/open-traffic-generator/snappi) against a back-to-back connected [Ixia-C](https://github.com/open-traffic-generator/ixia-c) topology.


### Prerequisites

- Recommended OS is Ubuntu LTS release.
- At least 2 CPU cores
- At least 6GB RAM
- At least 10GB Free Hard Disk Space
- Go 1.17+ or Python 3.6+ (with pip and pexpect)
- Docker Engine (Community Edition) ( https://docs.docker.com/engine/install/ubuntu/ )
- Docker Compose  ( sudo apt install docker-compose ) 
- 


### Usage:

1. Clone this repository

    ```sh
    git clone https://github.com/open-traffic-generator/conformance.git && cd conformance
    ```
    Go to the do.sh file and edit the GITHUB_USER and GITHUB_PAT variables before using this repo

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
    # run all feature tests against DP-only distribution of ixia-c
    ./do.sh gotest -tags="dp" ./feature/b2b/...
    # run all feature tests against CP/DP distribution of ixia-c
    ./do.sh gotest -tags="all" ./feature/b2b/...

    # run single test
    ./do.sh gotest ./feature/b2b/packet/udp/udp_port_value_test.go
    ```

4. Setup and run Python tests

    ```sh
    # setup test requirements
    ./do.sh prepytest

    # run all tests against DP-only distribution of ixia-c
    ./do.sh pytest -m dp ./feature/b2b/
    # run all tests against CP/DP distribution of ixia-c
    ./do.sh pytest ./feature/b2b/

    # run single test
    ./do.sh pytest ./feature/b2b/packet/udp/test_udp_port_value.py
    ```

5. Teardown topology

    ```sh
    # remove DP-only distribution of ixia-c
    ./do.sh topo rm dp

    # remove CP/DP distribution of ixia-c
    ./do.sh topo rm cpdp
    ```

### Advanced Usage:

1. Format python code
    
    Note that if you submit any code which does not follow proper python format the CI will fail

    ```sh
    # to format python code
    ./do.sh pylint

    # to format python code for a specific path
    ./do.sh pylint features
    ```

2. Format go code

   Note that if you submit any code which does not follow proper go format the CI will fail

    ```sh
    # to format go code
    ./do.sh golint

    # to format go code for a specific path
    ./do.sh golint helpers
    ```

3. Run perf tests in Go

    ```sh
    # run single perf test
    ./do.sh gotest ./performance/b2b/udp_mesh_flows_perf_test.go
    # run single perf test with lesser number of iterations (default=100)
    OTG_ITERATIONS=2 ./do.sh gotest ./performance/b2b/udp_mesh_flows_perf_test.go
    ```

4. Run tests against ixia-c B2B deployed on K8S cluster using eth0 as test interface

    ```sh
    # setup K8S cluster
    ./do.sh new_k8s_cluster
    # create topology
    ./do.sh topo new k8seth0
    # run single test
    ./do.sh gotest ./feature/b2b/packet/udp/udp_port_value_eth0_test.go
    # delete topology
    ./do.sh topo rm k8seth0
    # delete K8S cluster
    ./do.sh rm_k8s_cluster
    ```

5. Run tests against KNE cluster (Back-To-Back)

    ```sh
    # setup KNE cluster
    ./do.sh new_k8s_cluster kne
    # create KNE topology
    ./do.sh topo new kneb2b
    # run all back-to-back feature tests
    ./do.sh gotest -tags="all" ./feature/b2b/...
    # delete KNE topology
    ./do.sh topo rm kneb2b
    # delete KNE cluster
    ./do.sh rm_k8s_cluster kne
    ```

6. Run tests against KNE cluster (Port-Dut-Port)

    ```sh
    # setup KNE cluster
    ./do.sh new_k8s_cluster kne arista
    # create KNE topology
    ./do.sh topo new knepdp arista
    # run all port-dut-port feature tests
    ./do.sh gotest -tags="all" ./feature/pdp/...
    # delete KNE topology
    ./do.sh topo rm knepdp arista
    # delete KNE cluster
    ./do.sh rm_k8s_cluster kne arista
    ```

7. Run tests against OTG-HW (Back-To-Back)

    ```sh
    # create OTG-HW topology
    # user can change the Chassis ports to use "OTG_HW_PORT", inside the do.sh file.
    ./do.sh topo new otghw_b2b
    # run all back-to-back feature tests
    ./do.sh gotest -tags="all" ./feature/b2b/...
    # delete OTG-HW topology
    ./do.sh topo rm otghw_b2b
    ```

8. Run tests against UHD (Back-To-Back)

    Pre-requisite : check if the trunk interface ens192 link is up(command on terminal: ip link show). User can change the "TRUNK INTERFACE" in the do.sh. 
    The versions and path to images has been picked from versions.yaml file. User can change it as per their need.

    ```sh
    # create UHD topology-controller,gnmi and PEs. User can go to the do.sh file to change the "PE_LIST" to deploy and the "UHD_HOST" to Use For the Setup 
    # User can edit the test-config.yaml file after deployment to change the IP of the components to your machine IP as required, currently set as 10.36.87.205
    ./do.sh topo new uhd

    # deploy the uhd build you want to run on the uhd400 box. User can change the version and path of uhd, to be found on versions.yaml
    # change the "ip_address" and "password" inside the uhd_ssh.py file to your uhd box.
    python3 uhd_ssh.py

    #set all the uhd port 1-16 speed to 400 gbps and port 32 speed to 10 gbps.
    #Change the "UHD_IP" to your uhd box IP.
    #User can change the port speeds within the script file and run the command.
    ./set_uhd_port_speed.sh

    #check all the port metrics to see all configured correctly and link is up. Change the IP to your UHD IP. Intall jq before running this command using sudo apt install jq.
    curl -X POST -ksfL http://10.36.87.166/port/api/v1/monitor/metrics | jq '.port_metrics[] | {name, link}'

    # run back-to-back tests feature tests
    ./do.sh gotest -tags="all" ./feature/b2b/...

    # delete UHD topology
    ./do.sh topo rm uhd
    ```
