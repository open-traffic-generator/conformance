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

GO_VERSION=1.19
KIND_VERSION=v0.16.0
METALLB_VERSION=v0.13.6
MESHNET_COMMIT=f26c193
MESHNET_IMAGE="networkop/meshnet\:v0.3.0"
IXIA_C_OPERATOR_VERSION="0.3.0"
IXIA_C_OPERATOR_YAML="https://github.com/open-traffic-generator/ixia-c-operator/releases/download/v${IXIA_C_OPERATOR_VERSION}/ixiatg-operator.yaml"
KNE_COMMIT=a20cc6f

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

ixia_c_controller_img() {
    case $1 in
        dp  )
            ixia_c_img controller-dp
        ;;
        cpdp)
            ixia_c_img controller-cpdp
        ;;
        *   )
            echo "unsupported image type: ${1}"
            exit 1
        ;;
    esac
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
    && docker exec ixia-c-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml ixia-c-controller:${configdir}/ \
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
    && docker exec ixia-c-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml ixia-c-controller:${configdir}/ \
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
    && docker exec ixia-c-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml ixia-c-controller:${configdir}/ \
    && rm -rf ./config.yaml
}

gen_config_common() {
    location=localhost
    if [ "${1}" = "ipv6" ]
    then 
        location="[$(container_ip6 ixia-c-controller)]"
    fi

    yml="otg_host: https://${location}
        otg_speed: speed_1_gbps
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
    ADDR=$(kubectl get service -n ixia-c service-https-otg-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    # TODO: only works for B2B topology
    ETH1=$(grep a_int deployments/k8s/kne-manifests/${1}.yaml | cut -d\: -f2 | cut -d\  -f2)
    ETH2=$(grep z_int deployments/k8s/kne-manifests/${1}.yaml | cut -d\: -f2 | cut -d\  -f2)
    yml="otg_host: https://${ADDR}:8443
        otg_ports:
          - ${ETH1}
          - ${ETH2}
        "
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
        --name=ixia-c-controller                            \
        $(ixia_c_controller_img dp)                         \
        --accept-eula                                       \
        --debug                                             \
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
    docker stop ixia-c-controller && docker rm ixia-c-controller
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
        --name=ixia-c-controller                            \
        --publish 0.0.0.0:8443:8443                           \
        --publish 0.0.0.0:40051:40051                       \
        $(ixia_c_controller_img cpdp)                       \
        --accept-eula                                       \
        --debug                                             \
        --disable-app-usage-reporter                        \
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
    docker stop ixia-c-controller && docker rm ixia-c-controller

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
        --name=ixia-c-controller                            \
        --publish 0.0.0.0:8443:8443                           \
        --publish 0.0.0.0:40051:40051                       \
        $(ixia_c_controller_img cpdp)                       \
        --accept-eula                                       \
        --debug                                             \
        --disable-app-usage-reporter                        \
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

get_ixia_c_operator() {
    echo "Installing ixia-c-operator ${IXIA_C_OPERATOR_YAML} ..."
    kubectl apply -f ${IXIA_C_OPERATOR_YAML} \
    && wait_for_pods ixiatg-op-system
}

get_kne() {
    which kne_cli > /dev/null 2>&1 && return
    echo "Installing KNE ${KNE_COMMIT} ..."
    CGO_ENBLED=0 go install github.com/openconfig/kne/kne_cli@${KNE_COMMIT}
}

get_kubemod() {
    return
    kubectl label namespace kube-system admission.kubemod.io/ignore=true --overwrite \
    && kubectl apply -f https://raw.githubusercontent.com/kubemod/kubemod/v0.15.3/bundle.yaml \
    && wait_for_pods kubemod-system
}

setup_k8s_plugins() {
    echo "Setting up K8S plugins for ${1} ..."
    case $1 in
        kne  )
            get_metallb \
            && get_meshnet \
            && get_ixia_c_operator \
            && get_kne
        ;;
        *   )
            get_metallb
        ;;
    esac
}

ixia_c_image_path() {
    grep "\"${1}\"" -A 1 deployments/ixia-c-config.yaml | grep path | cut -d\" -f4
}

ixia_c_image_tag() {
    grep "\"${1}\"" -A 2 deployments/ixia-c-config.yaml | grep tag | cut -d\" -f4
}

load_ixia_c_images() {
    echo "Loading ixia-c images in cluster ..."
    login_ghcr
    for name in controller gnmi-server traffic-engine protocol-engine
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
    && setup_k8s_plugins ${1} \
    && load_ixia_c_images
}

rm_k8s_cluster() {
    echo "Deleting kind cluster ..."
    kind delete cluster 2> /dev/null
    rm -rf $HOME/.kube
}

kne_namespace() {
    grep -E "^name" deployments/k8s/kne-manifests/${1}.yaml | cut -d\  -f2 | sed -e s/\"//g
}
 
create_ixia_c_kne() {
    echo "Creating KNE ${1} topology ..."
    ns=$(kne_namespace ${1})
    kubectl apply -f deployments/ixia-c-config.yaml \
    && kne_cli create deployments/k8s/kne-manifests/${1}.yaml \
    && wait_for_pods ${ns} \
    && kubectl get pods -A \
    && kubectl get services -A \
    && gen_config_kne ${1} \
    && echo "Successfully deployed !"
}

rm_ixia_c_kne() {
    echo "Removing KNE ${1} topology ..."
    ns=$(kne_namespace ${1})
    kne_cli delete deployments/k8s/kne-manifests/${1}.yaml \
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
                lag )
                    create_ixia_c_b2b_lag
                ;;
                kneb2b )
                    create_ixia_c_kne ixia-c-b2b
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
                lag )
                    rm_ixia_c_b2b_cpdp
                ;;
                kneb2b )
                    rm_ixia_c_kne ixia-c-b2b
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
            mkdir -p logs/ixia-c-controller
            docker cp ixia-c-controller:/home/ixia-c/controller/logs/ logs/ixia-c-controller
            docker cp ixia-c-controller:/home/ixia-c/controller/config/config.yaml logs/ixia-c-controller
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

pregotest() {
    go mod download \
    && echo "Successfully setup gotest !"
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

    CGO_ENABLED=0 go test -v -count=1 -p=1 -timeout 3600s ${@} ./... | tee ${log}

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

case $1 in
    *   )
        # shift positional arguments so that arg 2 becomes arg 1, etc.
        cmd=${1}
        shift 1
        ${cmd} ${@} || usage
    ;;
esac
