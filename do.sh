#!/bin/sh

# update for any release using
# curl -kLO https://github.com/open-traffic-generator/ixia-c/releases/download/v0.0.1-2994/versions.yaml
VERSIONS_YAML="versions.yaml"
VETH_A="veth-a"
VETH_Z="veth-z"

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
    
    sudo mkdir -p /var/run/netn
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

ixia_c_traffic_engine_img() {
    path="ghcr.io/open-traffic-generator/ixia-c-traffic-engine"
    echo "${path}:$(grep traffic-engine ${VERSIONS_YAML} | cut -d\  -f2)"
}

ixia_c_protocol_engine_img() {
    path="ghcr.io/open-traffic-generator/licensed/ixia-c-protocol-engine"
    echo "${path}:$(grep protocol-engine ${VERSIONS_YAML} | cut -d\  -f2)"
}

ixia_c_controller_img() {
    path="ghcr.io/open-traffic-generator/ixia-c-controller"
    if [ "$1" = "lic" ]
    then
        path="ghcr.io/open-traffic-generator/licensed/ixia-c-controller"
    fi
    echo "${path}:$(grep controller ${VERSIONS_YAML} | cut -d\  -f2)"
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

gen_config_b2b_free() {
    yml="otg_host: https://localhost
        otg_ports:
          - localhost:5555
          - localhost:5556
        otg_speed: speed_1_gbps
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null
}

gen_config_b2b_lic() {
    OTG_HOST=$(container_ip ixia-c-controller)
    OTG_PORTA=$(container_ip ixia-c-traffic-engine-${VETH_A})
    OTG_PORTZ=$(container_ip ixia-c-traffic-engine-${VETH_Z})

    yml="otg_host: https://${OTG_HOST}
        otg_ports:
          - ${OTG_PORTA}:5555+${OTG_PORTA}:50071
          - ${OTG_PORTZ}:5555+${OTG_PORTZ}:50071
        otg_speed: speed_1_gbps
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null
}

create_ixia_c_b2b_free() {
    echo "Setting up back-to-back with free distribution of ixia-c ..."
    create_veth_pair ${VETH_A} ${VETH_Z}                    \
    && docker run --net=host  -d                            \
        --name=ixia-c-controller                            \
        $(ixia_c_controller_img)                            \
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
    && gen_config_b2b_free                                  \
    && echo "Successfully deployed !"
}

rm_ixia_c_b2b_free() {
    echo "Tearing down back-to-back with free distribution of ixia-c ..."
    docker stop ixia-c-controller && docker rm ixia-c-controller
    docker stop ixia-c-traffic-engine-${VETH_A}
    docker rm ixia-c-traffic-engine-${VETH_A}
    docker stop ixia-c-traffic-engine-${VETH_Z}
    docker rm ixia-c-traffic-engine-${VETH_Z}
    docker ps -a
    rm_veth_pair veth-a veth-z
}

create_ixia_c_b2b_licensed() {
    echo "Setting up back-to-back with licensed distribution of ixia-c ..."
    login_ghcr                                              \
    && docker run -d                                        \
        --name=ixia-c-controller                            \
        $(ixia_c_controller_img lic)                        \
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
    && gen_config_b2b_lic                                   \
    && sleep 30 \
    && echo "Successfully deployed !"
}

rm_ixia_c_b2b_licensed() {
    echo "Tearing down back-to-back with licensed distribution of ixia-c ..."
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

topo() {
    case $1 in
        new )
            if [ "${2}" = "lic" ]
            then
                create_ixia_c_b2b_licensed
            else
                create_ixia_c_b2b_free
            fi
        ;;
        rm  )
            if [ "${2}" = "lic" ]
            then
                rm_ixia_c_b2b_licensed
            else
                rm_ixia_c_b2b_free
            fi
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

    CGO_ENABLED=0 go test -v -count=1 ${@} ./... | tee ${log}

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

case $1 in
    *   )
        # shift positional arguments so that arg 2 becomes arg 1, etc.
        cmd=${1}
        shift 1
        ${cmd} ${@} || echo "usage: $0 [name of any function in script]"
    ;;
esac
