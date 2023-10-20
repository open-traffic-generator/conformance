#!/bin/sh

# update for any release using
# curl -kLO https://github.com/open-traffic-generator/ixia-c/releases/download/v0.0.1-2994/versions.yaml
VERSIONS_YAML="versions.yaml"
VETH_A="veth-a"
VETH_Z="veth-z"
# additional member ports for LAG
VETH_B="veth-b"
VETH_C="veth-c"
VETH_X="veth-x"
VETH_Y="veth-y"
FORCE_CREATE_VETH=true

GO_VERSION=1.19
KIND_VERSION=v0.20.0
METALLB_VERSION=v0.13.11
MESHNET_COMMIT=d7c306c
MESHNET_IMAGE="networkop/meshnet\:v0.3.0"
KENG_OPERATOR_VERSION="0.3.13"
KENG_OPERATOR_YAML="https://github.com/open-traffic-generator/keng-operator/releases/download/v${KENG_OPERATOR_VERSION}/ixiatg-operator.yaml"
NOKIA_SRL_OPERATOR_VERSION="0.4.6"
NOKIA_SRL_OPERATOR_YAML="https://github.com/srl-labs/srl-controller/config/default?ref=v${NOKIA_SRL_OPERATOR_VERSION}"
ARISTA_CEOS_OPERATOR_VERSION="2.0.1"
ARISTA_CEOS_OPERATOR_YAML="https://github.com/aristanetworks/arista-ceoslab-operator/config/default?ref=v${ARISTA_CEOS_OPERATOR_VERSION}"
ARISTA_CEOS_VERSION="4.29.1F-29233963"
ARISTA_CEOS_IMAGE="ghcr.io/open-traffic-generator/ceos"
KNE_VERSION=v0.1.15

OPENCONFIG_MODELS_REPO=https://github.com/openconfig/public.git
OPENCONFIG_MODELS_COMMIT=5ca6a36

TIMEOUT_SECONDS=300
APT_GET_UPDATE=true

apt_update() {
    if [ "${APT_UPDATE}" = "true" ]
    then
        sudo apt-get update
        APT_GET_UPDATE=false
    fi
}

apt_install() {
    echo "Installing ${1} ..."
    apt_update \
    && sudo apt-get install -y --no-install-recommends ${1}
}

apt_install_curl() {
    curl --version > /dev/null 2>&1 && return
    apt_install curl
}

apt_install_vim() {
    dpkg -s vim > /dev/null 2>&1 && return
    apt_install vim
}

apt_install_git() {
    git version > /dev/null 2>&1 && return
    apt_install git
}

apt_install_lsb_release() {
    lsb_release -v > /dev/null 2>&1 && return
    apt_install lsb_release
}

apt_install_gnupg() {
    gpg -k > /dev/null 2>&1 && return
    apt_install gnupg
}

apt_install_ca_certs() {
    dpkg -s ca-certificates > /dev/null 2>&1 && return
    apt_install gnupg ca-certificates
}

apt_install_pkgs() {
    uname -a | grep -i linux > /dev/null 2>&1 || return 0
    apt_install_curl \
    && apt_install_vim \
    && apt_install_git \
    && apt_install_lsb_release \
    && apt_install_gnupg \
    && apt_install_ca_certs
}

get_go() {
    which go > /dev/null 2>&1 && return
    echo "Installing Go ${GO_VERSION} ..."
    # install golang per https://golang.org/doc/install#tarball
    curl -kL https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz | sudo tar -C /usr/local/ -xzf - \
    && echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> $HOME/.profile \
    && . $HOME/.profile \
    && go version
}

get_docker() {
    which docker > /dev/null 2>&1 && return
    echo "Installing docker ..."
    sudo apt-get remove docker docker-engine docker.io containerd runc 2> /dev/null

    curl -kfsSL https://download.docker.com/linux/ubuntu/gpg \
        | sudo gpg --batch --yes --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
        | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    sudo apt-get update \
    && sudo apt-get install -y docker-ce docker-ce-cli containerd.io
}

common_install() {
    apt_install_pkgs \
    && get_go \
    && get_docker \
    && sudo_docker
}

create_veth_pair() {
    if [ -z "${1}" ] || [ -z "${2}" ]
    then
        echo "usage: ${0} create_veth_pair <name1> <name2>"
        exit 1
    fi

    if [ "${FORCE_CREATE_VETH}" = "true" ]
    then
        rm_veth_pair ${1} ${2} 2> /dev/null
    fi

    sudo ip link add ${1} type veth peer name ${2} \
    && sudo ip link set ${1} up \
    && sudo ip link set ${2} up
}

rm_veth_pair() {
    if [ -z "${1}" ] || [ -z "${2}" ]
    then
        echo "usage: ${0} rm_veth_pair <name1> <name2>"
        exit 1
    fi
    sudo ip link delete ${1}
}

ipv6_enable_docker() {
    echo "{\"ipv6\": true, \"fixed-cidr-v6\": \"2001:db8:1::/64\"}" | sudo tee /etc/docker/daemon.json
    echo "$(sudo systemctl restart docker)"
}

push_ifc_to_container() {
    if [ -z "${1}" ] || [ -z "${2}" ]
    then
        echo "usage: ${0} push_ifc_to_container <ifc-name> <container-name>"
        exit 1
    fi

    cid=$(container_id ${2})
    cpid=$(container_pid ${2})
    echo "Changing namespace of ifc ${1} to container ID ${cid} pid ${cpid}"
    orgPath=/proc/${cpid}/ns/net
    newPath=/var/run/netns/${cid}
    
    sudo mkdir -p /var/run/netns
    echo "Creating symlink ${orgPath} -> ${newPath}"
    sudo ln -s ${orgPath} ${newPath} \
    && sudo ip link set ${1} netns ${cid} \
    && sudo ip netns exec ${cid} ip link set ${1} name ${1} \
    && sudo ip netns exec ${cid} ip -4 addr add 0/0 dev ${1} \
    && sudo ip netns exec ${cid} ip -4 link set ${1} up \
    && echo "Successfully changed namespace of ifc ${1}"

    sudo rm -rf ${newPath}
}

container_id() {
    docker inspect --format="{{json .Id}}" ${1} | cut -d\" -f 2
}

container_pid() {
    docker inspect --format="{{json .State.Pid}}" ${1} | cut -d\" -f 2
}

container_ip() {
    docker inspect --format="{{json .NetworkSettings.IPAddress}}" ${1} | cut -d\" -f 2
}

container_ip6() {
    docker inspect --format="{{json .NetworkSettings.GlobalIPv6Address}}" ${1} | cut -d\" -f 2
}

ixia_c_img() {
    path=$(grep -A 2 ${1} ${VERSIONS_YAML} | grep path | cut -d: -f2 | cut -d\  -f2)
    tag=$(grep -A 2 ${1} ${VERSIONS_YAML} | grep tag | cut -d: -f2 | cut -d\  -f2)
    echo "${path}:${tag}"
}

ixia_c_traffic_engine_img() {
    ixia_c_img traffic-engine
}

ixia_c_protocol_engine_img() {
    ixia_c_img protocol-engine
}

keng_controller_img() {
    ixia_c_img controller
}

keng_license_server_img() {
    ixia_c_img license-server
}

login_ghcr() {
    if [ -f "$HOME/.docker/config.json" ]
    then
        grep ghcr.io "$HOME/.docker/config.json" > /dev/null && return 0
    fi

    if [ -z "${GITHUB_USER}" ] || [ -z "${GITHUB_PAT}" ]
    then
        echo "Logging into docker repo ghcr.io (Please provide Github Username and PAT)"
        docker login ghcr.io
    else
        echo "Logging into docker repo ghcr.io"
        echo "${GITHUB_PAT}" | docker login -u"${GITHUB_USER}" --password-stdin ghcr.io
    fi
}

logout_ghcr() {
    docker logout ghcr.io
}

gen_controller_config_b2b_dp() {
    configdir=/home/ixia-c/controller/config

    wait_for_sock localhost 5555
    wait_for_sock localhost 5556

    yml="location_map:
          - location: ${VETH_A}
            endpoint: localhost:5555
          - location: ${VETH_Z}
            endpoint: localhost:5556
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./config.yaml > /dev/null \
    && docker exec keng-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml keng-controller:${configdir}/ \
    && rm -rf ./config.yaml
}

gen_controller_config_b2b_cpdp() {
    configdir=/home/ixia-c/controller/config
    if [ "${1}" = "ipv6" ]
    then 
        OTG_PORTA=$(container_ip6 ixia-c-traffic-engine-${VETH_A})
        OTG_PORTZ=$(container_ip6 ixia-c-traffic-engine-${VETH_Z})
    else
        OTG_PORTA=$(container_ip ixia-c-traffic-engine-${VETH_A})
        OTG_PORTZ=$(container_ip ixia-c-traffic-engine-${VETH_Z})
    fi

    wait_for_sock ${OTG_PORTA} 5555
    wait_for_sock ${OTG_PORTA} 50071
    wait_for_sock ${OTG_PORTZ} 5555
    wait_for_sock ${OTG_PORTZ} 50071

    if [ "${1}" = "ipv6" ]
    then 
        OTG_PORTA="[${OTG_PORTA}]"
        OTG_PORTZ="[${OTG_PORTZ}]"
    fi

    yml="location_map:
          - location: ${VETH_A}
            endpoint: \"${OTG_PORTA}:5555+${OTG_PORTA}:50071\"
          - location: ${VETH_Z}
            endpoint: \"${OTG_PORTZ}:5555+${OTG_PORTZ}:50071\"
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./config.yaml > /dev/null \
    && docker exec keng-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml keng-controller:${configdir}/ \
    && rm -rf ./config.yaml
}

gen_controller_config_b2b_lag() {
    configdir=/home/ixia-c/controller/config
    OTG_PORTA=$(container_ip ixia-c-traffic-engine-${VETH_A})
    OTG_PORTZ=$(container_ip ixia-c-traffic-engine-${VETH_Z})

    wait_for_sock ${OTG_PORTA} 5555
    wait_for_sock ${OTG_PORTA} 50071
    wait_for_sock ${OTG_PORTZ} 5555
    wait_for_sock ${OTG_PORTZ} 50071

    yml="location_map:
          - location: ${VETH_A}
            endpoint: ${OTG_PORTA}:5555;1+${OTG_PORTA}:50071
          - location: ${VETH_B}
            endpoint: ${OTG_PORTA}:5555;2+${OTG_PORTA}:50071
          - location: ${VETH_C}
            endpoint: ${OTG_PORTA}:5555;3+${OTG_PORTA}:50071
          - location: ${VETH_Z}
            endpoint: ${OTG_PORTZ}:5555;1+${OTG_PORTZ}:50071
          - location: ${VETH_Y}
            endpoint: ${OTG_PORTZ}:5555;2+${OTG_PORTZ}:50071
          - location: ${VETH_X}
            endpoint: ${OTG_PORTZ}:5555;3+${OTG_PORTZ}:50071
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./config.yaml > /dev/null \
    && docker exec keng-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml keng-controller:${configdir}/ \
    && rm -rf ./config.yaml
}

gen_config_common() {
    location=localhost
    if [ "${1}" = "ipv6" ]
    then 
        location="[$(container_ip6 keng-controller)]"
    fi

    yml="otg_speed: speed_1_gbps
        otg_capture_check: true
        otg_iterations: 100
        otg_grpc_transport: false
        "
    echo -n "$yml" | sed "s/^        //g" | tee -a ./test-config.yaml > /dev/null
}

gen_config_b2b_dp() {
    yml="otg_host: https://localhost:8443
        otg_ports:
          - ${VETH_A}
          - ${VETH_Z}
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null

    gen_config_common 
}

gen_config_b2b_cpdp() {
    yml="otg_host: https://localhost:8443
        otg_ports:
          - ${VETH_A}
          - ${VETH_Z}
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null

    gen_config_common $1
}

gen_config_b2b_lag() {
    yml="otg_host: https://localhost:8443
        otg_ports:
          - ${VETH_A}
          - ${VETH_B}
          - ${VETH_C}
          - ${VETH_Z}
          - ${VETH_Y}
          - ${VETH_X}
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null

    gen_config_common
}

gen_config_kne() {
    OTG_ADDR=$(kubectl get service -n ixia-c service-https-otg-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
    topo=$(kne_topo_file ${1} ${2})
    yml="        otg_ports:\n"
    for p in $(grep "_node: otg" ${topo} -A 1 | grep _int | sed s/\\s//g)
    do
        yml="${yml}          - $(echo ${p} | cut -d: -f2)\n"
    done
    if [ ! -z "${2}" ]
    then
        yml="${yml}        dut_configs:\n"

        DUT_ADDR=$(kubectl get service -n ixia-c service-dut -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        yml="${yml}          - name: dut\n"
        yml="${yml}            host: ${DUT_ADDR}\n"
        
        SSH_PORT=$(grep "\- name: dut" -A 30 ${topo} | grep -m 1 services -A 10 | grep 'name: ssh' -B 1 | head -n 1 | tr -d ' :')
        yml="${yml}            ssh_username: admin\n"
        # yml="${yml}            ssh_password: admin\n"
        yml="${yml}            ssh_port: ${SSH_PORT}\n"

        GNMI_PORT=$(grep "\- name: dut" -A 30 ${topo} | grep -m 1 services -A 10 | grep 'name: gnmi' -B 1 | head -n 1 | tr -d ' :')
        yml="${yml}            gnmi_username: admin\n"
        yml="${yml}            gnmi_password: admin\n"
        yml="${yml}            gnmi_port: ${GNMI_PORT}\n"
        yml="${yml}            interfaces:\n"
        for i in $(kne topology service ${topo} | grep interfaces -A 8 | grep -E 'peer_name:\s+\"otg' -A 3 -B 5 | grep ' name:'| tr -d ' ')
        do
            ifc=$(echo $i | cut -d\" -f2)
            yml="${yml}              - ${ifc}\n"
        done

        wait_for_sock ${DUT_ADDR} ${GNMI_PORT}
        wait_for_sock ${DUT_ADDR} ${SSH_PORT}
    fi

    yml="otg_host: https://${OTG_ADDR}:8443\n${yml}"
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null

    gen_config_common
}

gen_config_k8s() {
    ADDR=$(kubectl get service -n ixia-c service-otg-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    ETH1=$(grep "location:" deployments/k8s/manifests/${1}.yaml -m 2 | head -n1 | cut -d\: -f2 | cut -d\  -f2)
    ETH2=$(grep "location:" deployments/k8s/manifests/${1}.yaml -m 2 | tail -n1 | cut -d\: -f2 | cut -d\  -f2)
    yml="otg_host: https://${ADDR}:8443
        otg_ports:
          - ${ETH1}
          - ${ETH2}
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null

    gen_config_common

    gen_config_k8s_test_const ${1}
}

gen_config_k8s_test_const() {
    if [ "${1}" = "ixia-c-b2b-eth0" ]
    then
        rxUdpPort=7000
        txIp=$(kubectl get pod -n ixia-c otg-port-eth1 -o 'jsonpath={.status.podIP}')
        rxIp=$(kubectl get pod -n ixia-c otg-port-eth2 -o 'jsonpath={.status.podIP}')
        # send ping to flood arp table and extract gateway MAC
        kubectl exec -n ixia-c otg-port-eth1 -c otg-port-eth1-protocol-engine -- ping -c 1 ${rxIp}
        gatewayMac=$(kubectl exec -n ixia-c otg-port-eth1 -c otg-port-eth1-protocol-engine -- arp -a | head -n 1 | cut -d\  -f4)
        txMac=$(kubectl exec -n ixia-c otg-port-eth1 -c otg-port-eth1-protocol-engine -- ifconfig eth0 | grep ether | sed 's/  */_/g' | cut -d_ -f3)
        rxMac=$(kubectl exec -n ixia-c otg-port-eth2 -c otg-port-eth2-protocol-engine -- ifconfig eth0 | grep ether | sed 's/  */_/g' | cut -d_ -f3)
        # drop UDP packets on given dst port
        kubectl exec -n ixia-c otg-port-eth2 -c otg-port-eth2-protocol-engine -- apt-get install -y iptables
        kubectl exec -n ixia-c otg-port-eth2 -c otg-port-eth2-protocol-engine -- iptables -A INPUT -p udp --destination-port ${rxUdpPort} -j DROP

        yml="otg_test_const:
              txMac: ${txMac}
              rxMac: ${rxMac}
              gatewayMac: ${gatewayMac}
              txIp: ${txIp}
              rxIp: ${rxIp}
              rxUdpPort: ${rxUdpPort}
            "
        echo -n "$yml" | sed "s/^            //g" | tee -a ./test-config.yaml > /dev/null
    fi
}

gen_openconfig_models_sdk() {
    echo "Fetching openconfig models ..."
    rm -rf etc/public \
    && rm -rf helpers/dut/gnmi \
    && mkdir -p etc \
    && cd etc \
    && git clone ${OPENCONFIG_MODELS_REPO} \
    && cd public \
    && git checkout ${OPENCONFIG_MODELS_COMMIT} \
    && cd ..

    EXCLUDE_MODULES=ietf-interfaces,openconfig-bfd,openconfig-messages

    YANG_FILES="
        public/release/models/acl/openconfig-acl.yang
        public/release/models/acl/openconfig-packet-match.yang
        public/release/models/aft/openconfig-aft.yang
        public/release/models/aft/openconfig-aft-network-instance.yang
        public/release/models/ate/openconfig-ate-flow.yang
        public/release/models/ate/openconfig-ate-intf.yang
        public/release/models/bfd/openconfig-bfd.yang
        public/release/models/bgp/openconfig-bgp-policy.yang
        public/release/models/bgp/openconfig-bgp-types.yang
        public/release/models/interfaces/openconfig-if-aggregate.yang
        public/release/models/interfaces/openconfig-if-ethernet.yang
        public/release/models/interfaces/openconfig-if-ethernet-ext.yang
        public/release/models/interfaces/openconfig-if-ip-ext.yang
        public/release/models/interfaces/openconfig-if-ip.yang
        public/release/models/interfaces/openconfig-if-sdn-ext.yang
        public/release/models/interfaces/openconfig-interfaces.yang
        public/release/models/isis/openconfig-isis.yang
        public/release/models/lacp/openconfig-lacp.yang
        public/release/models/lldp/openconfig-lldp-types.yang
        public/release/models/lldp/openconfig-lldp.yang
        public/release/models/local-routing/openconfig-local-routing.yang
        public/release/models/mpls/openconfig-mpls-types.yang
        public/release/models/multicast/openconfig-pim.yang
        public/release/models/network-instance/openconfig-network-instance.yang
        public/release/models/openconfig-extensions.yang
        public/release/models/optical-transport/openconfig-terminal-device.yang
        public/release/models/optical-transport/openconfig-transport-types.yang
        public/release/models/ospf/openconfig-ospfv2.yang
        public/release/models/p4rt/openconfig-p4rt.yang
        public/release/models/platform/openconfig-platform-cpu.yang
        public/release/models/platform/openconfig-platform-ext.yang
        public/release/models/platform/openconfig-platform-fan.yang
        public/release/models/platform/openconfig-platform-integrated-circuit.yang
        public/release/models/platform/openconfig-platform-software.yang
        public/release/models/platform/openconfig-platform-transceiver.yang
        public/release/models/platform/openconfig-platform.yang
        public/release/models/policy-forwarding/openconfig-policy-forwarding.yang
        public/release/models/policy/openconfig-policy-types.yang
        public/release/models/qos/openconfig-qos-elements.yang
        public/release/models/qos/openconfig-qos-interfaces.yang
        public/release/models/qos/openconfig-qos-types.yang
        public/release/models/qos/openconfig-qos.yang
        public/release/models/rib/openconfig-rib-bgp.yang
        public/release/models/sampling/openconfig-sampling-sflow.yang
        public/release/models/segment-routing/openconfig-segment-routing-types.yang
        public/release/models/system/openconfig-system.yang
        public/release/models/types/openconfig-inet-types.yang
        public/release/models/types/openconfig-types.yang
        public/release/models/types/openconfig-yang-types.yang
        public/release/models/vlan/openconfig-vlan.yang
        public/third_party/ietf/iana-if-type.yang
        public/third_party/ietf/ietf-inet-types.yang
        public/third_party/ietf/ietf-interfaces.yang
        public/third_party/ietf/ietf-yang-types.yang
    "

    go install github.com/openconfig/ygnmi/app/ygnmi@v0.7.6 \
    && go install golang.org/x/tools/cmd/goimports@v0.5.0 \
    && ygnmi generator \
        --trim_module_prefix=openconfig \
        --exclude_modules="${EXCLUDE_MODULES}" \
        --base_package_path=github.com/open-traffic-generator/conformance/helpers/dut/gnmi \
        --output_dir=../helpers/dut/gnmi \
        --paths=public/release/models/...,public/third_party/ietf/... \
        --ignore_deviate_notsupported \
        ${YANG_FILES} \
    && cd .. \
    && find helpers/dut/gnmi -name "*.go" -exec goimports -w {} + \
    && find helpers/dut/gnmi -name "*.go" -exec gofmt -w -s {} +
}

wait_for_sock() {
    start=$SECONDS
    TIMEOUT_SECONDS=30
    if [ ! -z "${3}" ]
    then
        TIMEOUT_SECONDS=${3}
    fi
    echo "Waiting for ${1}:${2} to be ready (timeout=${TIMEOUT_SECONDS}s)..."
    while true
    do
        nc -z -v ${1} ${2} && return 0

        elapsed=$(( SECONDS - start ))
        if [ $elapsed -gt ${TIMEOUT_SECONDS} ]
        then
            echo "${1}:${2} to be ready after ${TIMEOUT_SECONDS}s"
            exit 1
        fi
        sleep 0.1
    done

}

create_ixia_c_b2b_dp() {
    echo "Setting up back-to-back with DP-only distribution of ixia-c ..."
    create_veth_pair ${VETH_A} ${VETH_Z}                    \
    && docker run --net=host  -d                            \
        --name=keng-controller                              \
        $(keng_controller_img)                              \
        --accept-eula                                       \
        --trace                                             \
        --disable-app-usage-reporter                        \
    && docker run --net=host --privileged -d                \
        --name=ixia-c-traffic-engine-${VETH_A}              \
        -e OPT_LISTEN_PORT="5555"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_A}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        $(ixia_c_traffic_engine_img)                        \
    && docker run --net=host --privileged -d                \
        --name=ixia-c-traffic-engine-${VETH_Z}              \
        -e OPT_LISTEN_PORT="5556"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_Z}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        $(ixia_c_traffic_engine_img)                        \
    && docker ps -a                                         \
    && gen_controller_config_b2b_dp                         \
    && gen_config_b2b_dp                                    \
    && echo "Successfully deployed !"
}

rm_ixia_c_b2b_dp() {
    echo "Tearing down back-to-back with DP-only distribution of ixia-c ..."
    docker stop keng-controller && docker rm keng-controller
    docker stop ixia-c-traffic-engine-${VETH_A}
    docker rm ixia-c-traffic-engine-${VETH_A}
    docker stop ixia-c-traffic-engine-${VETH_Z}
    docker rm ixia-c-traffic-engine-${VETH_Z}
    docker ps -a
    rm_veth_pair ${VETH_A} ${VETH_Z}
}

create_ixia_c_b2b_cpdp() {
    if [ "${1}" = "ipv6" ]
    then 
        ipv6_enable_docker
    fi
    echo "Setting up back-to-back with CP/DP distribution of ixia-c ..."
    login_ghcr                                              \
    && docker run -d                                        \
        --name=keng-controller                              \
        --publish 0.0.0.0:8443:8443                         \
        --publish 0.0.0.0:40051:40051                       \
        $(keng_controller_img)                              \
        --accept-eula                                       \
        --trace                                             \
        --disable-app-usage-reporter                        \
        --license-servers localhost                         \
    && docker run -d                                        \
        --net=container:keng-controller                     \
        --name=keng-license-server                          \
        $(keng_license_server_img)                          \
        --accept-eula                                       \
        --debug                                             \
    && docker run --privileged -d                           \
        --name=ixia-c-traffic-engine-${VETH_A}              \
        -e OPT_LISTEN_PORT="5555"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_A}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        -e WAIT_FOR_IFACE="Yes"                             \
        $(ixia_c_traffic_engine_img)                        \
    && docker run --privileged -d                           \
        --net=container:ixia-c-traffic-engine-${VETH_A}     \
        --name=ixia-c-protocol-engine-${VETH_A}             \
        -e INTF_LIST="${VETH_A}"                            \
        $(ixia_c_protocol_engine_img)                       \
    && docker run --privileged -d                           \
        --name=ixia-c-traffic-engine-${VETH_Z}              \
        -e OPT_LISTEN_PORT="5555"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_Z}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        -e WAIT_FOR_IFACE="Yes"                             \
        $(ixia_c_traffic_engine_img)                        \
    && docker run --privileged -d                           \
        --net=container:ixia-c-traffic-engine-${VETH_Z}     \
        --name=ixia-c-protocol-engine-${VETH_Z}             \
        -e INTF_LIST="${VETH_Z}"                            \
        $(ixia_c_protocol_engine_img)                       \
    && docker ps -a                                         \
    && create_veth_pair ${VETH_A} ${VETH_Z}                 \
    && push_ifc_to_container ${VETH_A} ixia-c-traffic-engine-${VETH_A}  \
    && push_ifc_to_container ${VETH_Z} ixia-c-traffic-engine-${VETH_Z}  \
    && gen_controller_config_b2b_cpdp $1                     \
    && gen_config_b2b_cpdp $1                                \
    && docker ps -a \
    && echo "Successfully deployed !"
}

rm_ixia_c_b2b_cpdp() {
    echo "Tearing down back-to-back with CP/DP distribution of ixia-c ..."
    docker stop keng-controller && docker rm keng-controller

    docker stop ixia-c-traffic-engine-${VETH_A}
    docker stop ixia-c-protocol-engine-${VETH_A}
    docker rm ixia-c-traffic-engine-${VETH_A}
    docker rm ixia-c-protocol-engine-${VETH_A}

    docker stop ixia-c-traffic-engine-${VETH_Z}
    docker stop ixia-c-protocol-engine-${VETH_Z}
    docker rm ixia-c-traffic-engine-${VETH_Z}
    docker rm ixia-c-protocol-engine-${VETH_Z}

    docker ps -a
}

create_ixia_c_b2b_lag() {
    echo "Setting up back-to-back LAG with CP/DP distribution of ixia-c ..."
    login_ghcr                                              \
    && docker run -d                                        \
        --name=keng-controller                              \
        --publish 0.0.0.0:8443:8443                         \
        --publish 0.0.0.0:40051:40051                       \
        $(keng_controller_img)                              \
        --accept-eula                                       \
        --trace                                             \
        --disable-app-usage-reporter                        \
        --license-servers localhost                         \
    && docker run -d                                        \
        --net=container:keng-controller                     \
        --name=keng-license-server                          \
        $(keng_license_server_img)                          \
        --accept-eula                                       \
        --debug                                             \
    && docker run --privileged -d                           \
        --name=ixia-c-traffic-engine-${VETH_A}              \
        -e OPT_LISTEN_PORT="5555"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_A} virtual@af_packet,${VETH_B} virtual@af_packet,${VETH_C}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        -e WAIT_FOR_IFACE="Yes"                             \
        -e OPT_MEMORY="1024"                                \
        $(ixia_c_traffic_engine_img)                        \
    && docker run --privileged -d                           \
        --net=container:ixia-c-traffic-engine-${VETH_A}     \
        --name=ixia-c-protocol-engine-${VETH_A}             \
        -e INTF_LIST="${VETH_A},${VETH_B},${VETH_C}"        \
        $(ixia_c_protocol_engine_img)                       \
    && docker run --privileged -d                           \
        --name=ixia-c-traffic-engine-${VETH_Z}              \
        -e OPT_LISTEN_PORT="5555"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_Z} virtual@af_packet,${VETH_Y} virtual@af_packet,${VETH_X}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        -e WAIT_FOR_IFACE="Yes"                             \
        -e OPT_MEMORY="1024"                                \
        $(ixia_c_traffic_engine_img)                        \
    && docker run --privileged -d                           \
        --net=container:ixia-c-traffic-engine-${VETH_Z}     \
        --name=ixia-c-protocol-engine-${VETH_Z}             \
        -e INTF_LIST="${VETH_Z},${VETH_Y},${VETH_X}"        \
        $(ixia_c_protocol_engine_img)                       \
    && docker ps -a                                         \
    && create_veth_pair ${VETH_A} ${VETH_Z}                 \
    && create_veth_pair ${VETH_B} ${VETH_Y}                 \
    && create_veth_pair ${VETH_C} ${VETH_X}                 \
    && push_ifc_to_container ${VETH_A} ixia-c-traffic-engine-${VETH_A}  \
    && push_ifc_to_container ${VETH_Z} ixia-c-traffic-engine-${VETH_Z}  \
    && push_ifc_to_container ${VETH_B} ixia-c-traffic-engine-${VETH_A}  \
    && push_ifc_to_container ${VETH_Y} ixia-c-traffic-engine-${VETH_Z}  \
    && push_ifc_to_container ${VETH_C} ixia-c-traffic-engine-${VETH_A}  \
    && push_ifc_to_container ${VETH_X} ixia-c-traffic-engine-${VETH_Z}  \
    && gen_controller_config_b2b_lag                        \
    && gen_config_b2b_lag                                   \
    && docker ps -a \
    && echo "Successfully deployed !"
}

sudo_docker() {
    groups | grep docker > /dev/null 2>&1 && return
    sudo groupadd docker
    sudo usermod -aG docker $USER

    sudo docker version
    echo "Please logout, login again and re-execute previous command"
    exit 0
}

get_kind() {
    which kind > /dev/null 2>&1 && return
    echo "Installing kind ${KIND_VERSION} ..."
    go install sigs.k8s.io/kind@${KIND_VERSION}
}

kind_cluster_exists() {
    kind get clusters | grep kind > /dev/null 2>&1
}

kind_create_cluster() {
    kind_cluster_exists && return
    echo "Creating kind cluster ..."
    kind create cluster --config=deployments/k8s/kind.yaml --wait ${TIMEOUT_SECONDS}s
}

kind_get_kubectl() {
    echo "Copying kubectl from kind cluster to host ..."
    rm -rf kubectl
    docker cp kind-control-plane:/usr/bin/kubectl ./ \
    && sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl \
    && sudo cp -r $HOME/.kube /root/ \
    && rm -rf kubectl
}

setup_kind_cluster() {
    echo "Setting up kind cluster ..."
    get_kind \
    && kind_create_cluster \
    && kind_get_kubectl
}

mk_metallb_config() {
    prefix=$(docker network inspect -f '{{.IPAM.Config}}' kind | grep -Eo "[0-9]+\.[0-9]+\.[0-9]+" | tail -n 1)

    yml="apiVersion: metallb.io/v1beta1
        kind: IPAddressPool
        metadata:
          name: kne-pool
          namespace: metallb-system
        spec:
          addresses:
            - ${prefix}.100 - ${prefix}.250

        ---
        apiVersion: metallb.io/v1beta1
        kind: L2Advertisement
        metadata:
          name: kne-l2-adv
          namespace: metallb-system
        spec:
          ipAddressPools:
            - kne-pool
    "

    echo "$yml" | sed "s/^        //g" | tee deployments/k8s/metallb.yaml > /dev/null
}

get_metallb() {
    echo "Setting up metallb ..."
    kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/${METALLB_VERSION}/config/manifests/metallb-native.yaml \
    && wait_for_pods metallb-system \
    && mk_metallb_config \
    && echo "Applying metallb config map for exposing internal services via public IP addresses ..." \
    && cat deployments/k8s/metallb.yaml \
    && kubectl apply -f deployments/k8s/metallb.yaml
}

get_meshnet() {
    echo "Installing meshnet-cni (${MESHNET_COMMIT}) ..."
    rm -rf deployments/k8s/meshnet-cni
    oldpwd=${PWD}
    cd deployments/k8s

    git clone https://github.com/networkop/meshnet-cni && cd meshnet-cni && git checkout ${MESHNET_COMMIT} \
    && cat manifests/base/daemonset.yaml | sed "s#image: networkop/meshnet:latest#image: ${MESHNET_IMAGE}#g" | tee manifests/base/daemonset.yaml.patched > /dev/null \
    && mv manifests/base/daemonset.yaml.patched manifests/base/daemonset.yaml \
    && kubectl apply -k manifests/base \
    && wait_for_pods meshnet \
    && cd ${oldpwd}
}

get_keng_operator() {
    echo "Installing keng-operator ${KENG_OPERATOR_YAML} ..."
    load_image_to_kind $(keng_operator_image) \
    && kubectl apply -f ${KENG_OPERATOR_YAML} \
    && wait_for_pods ixiatg-op-system
}

rm_keng_operator() {
    echo "Removing keng-operator ${KENG_OPERATOR_YAML} ..."
    kubectl delete -f ${KENG_OPERATOR_YAML} \
    && wait_for_no_namespace ixiatg-op-system
}

get_nokia_srl_operator() {
    echo "Installing nokia srl operator ${NOKIA_SRL_OPERATOR_YAML} ..."
    load_image_to_kind $(nokia_srl_operator_image) "local" \
    && kubectl apply -k ${NOKIA_SRL_OPERATOR_YAML} \
    && wait_for_pods srlinux-controller
}

rm_nokia_srl_operator() {
    echo "Removing nokia srl operator ${NOKIA_SRL_OPERATOR_YAML} ..."
    kubectl delete -k ${NOKIA_SRL_OPERATOR_YAML} \
    && wait_for_no_namespace srlinux-controller
}

get_arista_ceos_operator() {
    echo "Installing arista ceos operator ${ARISTA_CEOS_OPERATOR_YAML} ..."
    load_image_to_kind $(arista_ceos_operator_image) "local" \
    && kubectl apply -k ${ARISTA_CEOS_OPERATOR_YAML} \
    && wait_for_pods arista-ceoslab-operator-system
}

rm_arista_ceos_operator() {
    echo "Removing arista ceos operator ${ARISTA_CEOS_OPERATOR_YAML} ..."
    kubectl delete -k ${ARISTA_CEOS_OPERATOR_YAML} \
    && wait_for_no_namespace arista-ceoslab-operator-system
}

get_kne() {
    which kne > /dev/null 2>&1 && return
    echo "Installing KNE ${KNE_VERSION} ..."
    CGO_ENBLED=0 go install github.com/openconfig/kne/kne_cli@${KNE_VERSION} \
    && mv $(which kne_cli) $(dirname $(which kne_cli))/kne
}

get_kubemod() {
    return
    kubectl label namespace kube-system admission.kubemod.io/ignore=true --overwrite \
    && kubectl apply -f https://raw.githubusercontent.com/kubemod/kubemod/v0.15.3/bundle.yaml \
    && wait_for_pods kubemod-system
}

setup_k8s_plugins() {
    echo "Setting up K8S plugins for ${1} ${2} ..."
    case $1 in
        kne  )
            get_metallb \
            && get_meshnet \
            && get_keng_operator \
            && get_kne
        ;;
        *   )
            get_metallb
        ;;
    esac

    case $2 in
        nokia   )
            get_nokia_srl_operator
        ;;
        arista  )
            get_arista_ceos_operator \
            && load_arista_ceos_image
        ;;
        *       )
            echo "second argument '${2}' ignored"
        ;;
    esac
}

ixia_c_image_path() {
    grep "\"${1}\"" -A 1 deployments/ixia-c-config.yaml | grep path | cut -d\" -f4
}

ixia_c_image_tag() {
    grep "\"${1}\"" -A 2 deployments/ixia-c-config.yaml | grep tag | cut -d\" -f4
}

keng_operator_image() {
    curl -kLs ${KENG_OPERATOR_YAML} | grep image | grep operator | tr -s ' ' | cut -d\  -f3
}

nokia_srl_operator_image() {
    yml="$(curl -kLs https://raw.githubusercontent.com/srl-labs/srl-controller/v${NOKIA_SRL_OPERATOR_VERSION}/config/manager/kustomization.yaml)"
    path=$(echo "${yml}" | grep newName | tr -s ' ' | cut -d\  -f 3)
    tag=$(echo "${yml}" | grep newTag | tr -s ' ' | cut -d\  -f 3)
    echo "${path}:${tag}"
}

arista_ceos_operator_image() {
    yml="$(curl -kLs https://raw.githubusercontent.com/aristanetworks/arista-ceoslab-operator/v${ARISTA_CEOS_OPERATOR_VERSION}/config/manager/kustomization.yaml)"
    path=$(echo "${yml}" | grep newName | tr -s ' ' | cut -d\  -f 3)
    tag=$(echo "${yml}" | grep newTag | tr -s ' ' | cut -d\  -f 3)
    echo "${path}:${tag}"
}

load_image_to_kind() {
    echo "Loading ${1}"

    login_ghcr \
    && docker pull "${1}" \
    && kind load docker-image "${1}" \
    || return 1

    if [ "${2}" = "local" ]
    then
        localimg="$(echo ${1} | cut -d: -f1):local"
        docker tag "${1}" "${localimg}" \
        && kind load docker-image "${localimg}"
    fi
}

load_arista_ceos_image() {
    load_image_to_kind "${ARISTA_CEOS_IMAGE}:${ARISTA_CEOS_VERSION}" "local"
}

load_ixia_c_images() {
    echo "Loading ixia-c images in cluster ..."
    login_ghcr
    for name in controller gnmi-server traffic-engine protocol-engine license-server
    do
        p=$(ixia_c_image_path ${name})
        t=$(ixia_c_image_tag ${name})
        img=${p}:${t}
        limg=${p}:local
        echo "Loading ${img}"
        docker pull ${img} \
        && docker tag ${img} ${limg} \
        && kind load docker-image ${img} \
        && kind load docker-image ${limg}
    done
}

wait_for_pods() {
    for n in $(kubectl get namespaces -o 'jsonpath={.items[*].metadata.name}')
    do
        if [ ! -z "$1" ] && [ "$1" != "$n" ]
        then
            continue
        fi
        for p in $(kubectl get pods -n ${n} -o 'jsonpath={.items[*].metadata.name}')
        do
            if [ ! -z "$2" ] && [ "$2" != "$p" ]
            then
                continue
            fi
            echo "Waiting for pod/${p} in namespace ${n} (timeout=${TIMEOUT_SECONDS}s)..."
            kubectl wait -n ${n} pod/${p} --for condition=ready --timeout=${TIMEOUT_SECONDS}s
        done
    done
}

wait_for_no_namespace() {
    start=$SECONDS
    echo "Waiting for namespace ${1} to be deleted (timeout=${TIMEOUT_SECONDS}s)..."
    while true
    do
        found=""
        for n in $(kubectl get namespaces -o 'jsonpath={.items[*].metadata.name}')
        do
            if [ "$1" = "$n" ]
            then
                found="$n"
                break
            fi
        done

        if [ -z "$found" ]
        then
            return 0
        fi

        elapsed=$(( SECONDS - start ))
        if [ $elapsed -gt ${TIMEOUT_SECONDS} ]
        then
            err "Namespace ${1} not deleted after ${TIMEOUT_SECONDS}s" 1
        fi
    done
}

new_k8s_cluster() {
    common_install \
    && setup_kind_cluster \
    && setup_k8s_plugins ${1} ${2} \
    && load_ixia_c_images
}

rm_k8s_cluster() {
    echo "Deleting kind cluster ..."
    kind delete cluster 2> /dev/null
    rm -rf $HOME/.kube
}

kne_topo_file() {
    path=deployments/k8s/kne-manifests
    if [ -z "${2}" ]
    then
        echo ${path}/${1}.yaml
    else
        echo ${path}/${1}-${2}.yaml
    fi
}

kne_namespace() {
    grep -E "^name" $(kne_topo_file ${1} ${2}) | cut -d\  -f2 | sed -e s/\"//g
}
 
create_ixia_c_kne() {
    echo "Creating KNE ${1} ${2} topology ..."
    ns=$(kne_namespace ${1} ${2})
    topo=$(kne_topo_file ${1} ${2})
    kubectl apply -f deployments/ixia-c-config.yaml \
    && kne create ${topo} \
    && wait_for_pods ${ns} \
    && kubectl get pods -A \
    && kubectl get services -A \
    && gen_config_kne ${1} ${2} \
    && echo "Successfully deployed !"
}

rm_ixia_c_kne() {
    echo "Removing KNE ${1} topology ..."
    ns=$(kne_namespace ${1} ${2})
    topo=$(kne_topo_file ${1} ${2})
    kne delete ${topo} \
    && wait_for_no_namespace ${ns}
}

k8s_namespace() {
    grep namespace deployments/k8s/manifests/${1}.yaml -m 1 | cut -d \: -f2 | cut -d \  -f 2
}

create_ixia_c_k8s() {
    echo "Creating K8S ${1} topology ..."
    ns=$(k8s_namespace ${1})
    kubectl apply -f deployments/k8s/manifests/${1}.yaml \
    && wait_for_pods ${ns} \
    && kubectl get pods -A \
    && kubectl get services -A \
    && gen_config_k8s ${1} \
    && echo "Successfully deployed !"
}

rm_ixia_c_k8s() {
    echo "Removing K8S ${1} topology ..."
    ns=$(k8s_namespace ${1})
    kubectl delete -f deployments/k8s/manifests/${1}.yaml \
    && wait_for_no_namespace ${ns}
}

topo() {
    case $1 in
        new )
            case $2 in
                dp  )
                    create_ixia_c_b2b_dp
                ;;
                cpdp)
                    create_ixia_c_b2b_cpdp $3
                ;;
                b2blag )
                    create_ixia_c_b2b_lag
                ;;
                kneb2b )
                    create_ixia_c_kne ixia-c-b2b
                ;;
                knepdp )
                    create_ixia_c_kne ixia-c-pdp ${3}
                ;;
                k8seth0 )
                    create_ixia_c_k8s ixia-c-b2b-eth0
                ;;
                *   )
                    echo "unsupported topo type: ${2}"
                    exit 1
                ;;
            esac
        ;;
        rm  )
            case $2 in
                dp  )
                    rm_ixia_c_b2b_dp
                ;;
                cpdp)
                    rm_ixia_c_b2b_cpdp
                ;;
                b2blag )
                    rm_ixia_c_b2b_cpdp
                ;;
                kneb2b )
                    rm_ixia_c_kne ixia-c-b2b
                ;;
                knepdp )
                    rm_ixia_c_kne ixia-c-pdp ${3}
                ;;
                k8seth0 )
                    rm_ixia_c_k8s ixia-c-b2b-eth0
                ;;
                *   )
                    echo "unsupported topo type: ${2}"
                    exit 1
                ;;
            esac
        ;;
        logs    )
            mkdir -p logs/keng-controller
            docker cp keng-controller:/home/ixia-c/controller/logs/ logs/keng-controller
            docker cp keng-controller:/home/ixia-c/controller/config/config.yaml logs/keng-controller
            mkdir -p logs/ixia-c-traffic-engine-${VETH_A}
            mkdir -p logs/ixia-c-traffic-engine-${VETH_Z}
            docker cp ixia-c-traffic-engine-${VETH_A}:/var/log/usstream/ logs/ixia-c-traffic-engine-${VETH_A}
            docker cp ixia-c-traffic-engine-${VETH_Z}:/var/log/usstream/ logs/ixia-c-traffic-engine-${VETH_Z}
            if [ "${2}" = "cpdp" ]
            then
                mkdir -p logs/ixia-c-protocol-engine-${VETH_A}
                mkdir -p logs/ixia-c-protocol-engine-${VETH_Z}
                docker cp ixia-c-protocol-engine-${VETH_A}:/var/log/ logs/ixia-c-protocol-engine-${VETH_A}
                docker cp ixia-c-protocol-engine-${VETH_Z}:/var/log/ logs/ixia-c-protocol-engine-${VETH_Z}
                # TODO: where to get complete logs ?
                docker logs ixia-c-protocol-engine-${VETH_A} | tee logs/ixia-c-protocol-engine-${VETH_A}/stdout.log > /dev/null
                docker logs ixia-c-protocol-engine-${VETH_Z} | tee logs/ixia-c-protocol-engine-${VETH_Z}/stdout.log > /dev/null
            fi
            top -bn2 | tee logs/resource-usage.log > /dev/null
        ;;
        *   )
            exit 1
        ;;
    esac
}

prepytest() {
    rm -rf .env
    python -m pip install virtualenv \
    && python -m virtualenv .env \
    && .env/bin/python -m pip install -r requirements.txt \
    && echo "Successfully setup pytest !"
}

gotest() {
    mkdir -p logs
    log=logs/gotest.log

    CGO_ENABLED=0 go test -v -count=1 -p=1 -timeout 3600s ${@} | tee ${log}

    echo "Summary:"
    grep ": Test" ${log}

    grep FAIL ${log} && return 1 || true
}

pytest() {
    mkdir -p logs
    py=.env/bin/python
    log=logs/pytest.log

    ${py} -m pytest -svvv ${@} | tee ${log}
    
    grep FAILED ${log} && return 1 || true
}


help() {
    grep "() {" ${0} | cut -d\  -f1
}

usage() {
    echo "usage: $0 [name of any function in script]"
    exit 1
}

pylint() {
    mkdir -p logs
    py=.env/bin/python
    log=logs/pylint.log

    lintdir=$([ -z "${1}" ] && echo "." || echo ${1})

    if [ -z "${CI}" ]
    then
        ${py} -m black ${lintdir}
    else
        ${py} -m black ${lintdir} --check > ${log} 2>&1
        sed 's/would reformat/Black formatting failed for/' ${log} 
    fi
}

golint() {

    GO111MODULE=on CGO_ENABLED=0 go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
    # TODO: skip-dirs does not actually skip analysis, it just supresses warnings
    if [ -z "${CI}" ]
    then
        fmtdir=$([ -z "${1}" ] && echo "." || echo ${1})
        gofmt -s -w ${fmtdir} \
        && echo "files reformatted"

        lintdir=$([ -z "${1}" ] && echo "./..." || echo ${1})
        golangci-lint run --timeout 30m -v ${lintdir} --skip-dirs helpers/dut/gnmi
    else
        lintdir=$([ -z "${1}" ] && echo "./..." || echo ${1})
        golangci-lint run --timeout 30m -v ${lintdir} --skip-dirs helpers/dut/gnmi
    fi 

}

case $1 in
    *   )
        # shift positional arguments so that arg 2 becomes arg 1, etc.
        cmd=${1}
        shift 1
        ${cmd} ${@} || usage
    ;;
esac
