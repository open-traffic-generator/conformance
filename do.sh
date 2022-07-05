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
    if [ "$1" = "licensed" ]
    then
        path="path=ghcr.io/open-traffic-generator/licensed/ixia-c-controller"
    fi
    echo "${path}:$(grep controller ${VERSIONS_YAML} | cut -d\  -f2)"
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
        --name=ixia-c-traffic-engine-a                      \
        -e OPT_LISTEN_PORT="5555"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_A}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        $(ixia_c_traffic_engine_img)                        \
    && docker run --net=host --privileged -d                \
        --name=ixia-c-traffic-engine-z                      \
        -e OPT_LISTEN_PORT="5556"                           \
        -e ARG_IFACE_LIST="virtual@af_packet,${VETH_Z}"     \
        -e OPT_NO_HUGEPAGES="Yes"                           \
        -e OPT_NO_PINNING="Yes"                             \
        $(ixia_c_traffic_engine_img)                        \
    && docker ps -a                                         \
    && echo "Successfully deployed !"
}

rm_ixia_c_b2b_free() {
    echo "Tearing down back-to-back with free distribution of ixia-c ..."
    docker stop ixia-c-controller && docker rm ixia-c-controller
    docker stop ixia-c-traffic-engine-a && docker rm ixia-c-traffic-engine-a
    docker stop ixia-c-traffic-engine-z && docker rm ixia-c-traffic-engine-z
    docker ps -a
    rm_veth_pair veth-a veth-z
}

create_ixia_c_b2b_licensed() {
    echo "Setting up back-to-back with licensed distribution of ixia-c ..."
    docker login ghcr.io
    create_veth_pair veth-a veth-z
}

rm_ixia_c_b2b_licensed() {
    echo "Tearing down back-to-back with licensed distribution of ixia-c ..."
    rm_veth_pair veth-a veth-z
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

    # TODO: path should be ./... instead of ./flows/... (but ./... doesn't stream log output)
    if [ -z ${1} ]
    then
        CGO_ENABLED=0 go test -v -count=1 ./flows/... | tee ${log}
    else
        CGO_ENABLED=0 go test -v -count=1 -run "^${1}$" ./flows/... | tee ${log}
    fi
    
    echo "Summary:"
    grep ": Test" ${log}

    grep FAIL ${log} && return 1 || true
}

pytest() {
    mkdir -p logs
    py=.env/bin/python
    log=logs/pytest.log

    if [ -z ${1} ]
    then
        ${py} -m pytest -svvv | tee ${log}
    else
        ${py} -m pytest -svvv -k "${1}" | tee ${log}
    fi
    
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
