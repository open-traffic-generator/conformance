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
enable_ipv6=false

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
    if [ ${enable_ipv6} = true ]
    then
        docker inspect --format="{{json .NetworkSettings.GlobalIPv6Address}}" ${1} | cut -d\" -f 2
    else
        docker inspect --format="{{json .NetworkSettings.IPAddress}}" ${1} | cut -d\" -f 2
    fi
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
    configdir=/home/keysight/ixia-c/controller/config

    wait_for_sock localhost 5555
    wait_for_sock localhost 5556

    yml="location_map:
          - location: ${VETH_A}
            endpoint: \"[localhost]:5555\"
          - location: ${VETH_Z}
            endpoint: \"[localhost]:5556\"
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./config.yaml > /dev/null \
    && docker exec ixia-c-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml ixia-c-controller:${configdir}/ \
    && rm -rf ./config.yaml
}

gen_controller_config_b2b_cpdp() {
    configdir=/home/keysight/ixia-c/controller/config
    OTG_PORTA=$(container_ip ixia-c-traffic-engine-${VETH_A})
    OTG_PORTZ=$(container_ip ixia-c-traffic-engine-${VETH_Z})

    wait_for_sock ${OTG_PORTA} 5555
    wait_for_sock ${OTG_PORTA} 50071
    wait_for_sock ${OTG_PORTZ} 5555
    wait_for_sock ${OTG_PORTZ} 50071

    yml="location_map:
          - location: ${VETH_A}
            endpoint: \"[${OTG_PORTA}]:5555+[${OTG_PORTA}]:50071\"
          - location: ${VETH_Z}
            endpoint: \"[${OTG_PORTZ}]:5555+[${OTG_PORTZ}]:50071\"
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./config.yaml > /dev/null \
    && docker exec ixia-c-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml ixia-c-controller:${configdir}/ \
    && rm -rf ./config.yaml
}

gen_controller_config_b2b_lag() {
    configdir=/home/keysight/ixia-c/controller/config
    OTG_PORTA=$(container_ip ixia-c-traffic-engine-${VETH_A})
    OTG_PORTZ=$(container_ip ixia-c-traffic-engine-${VETH_Z})

    wait_for_sock ${OTG_PORTA} 5555
    wait_for_sock ${OTG_PORTA} 50071
    wait_for_sock ${OTG_PORTZ} 5555
    wait_for_sock ${OTG_PORTZ} 50071

    yml="location_map:
          - location: ${VETH_A}
            endpoint: \"[${OTG_PORTA}]:5555;1+[${OTG_PORTA}]:50071\"
          - location: ${VETH_B}
            endpoint: \"[${OTG_PORTA}]:5555;2+[${OTG_PORTA}]:50071\"
          - location: ${VETH_C}
            endpoint: \"[${OTG_PORTA}]:5555;3+[${OTG_PORTA}]:50071\"
          - location: ${VETH_Z}
            endpoint: \"[${OTG_PORTZ}]:5555;1+[${OTG_PORTZ}]:50071\"
          - location: ${VETH_Y}
            endpoint: \"[${OTG_PORTZ}]:5555;2+[${OTG_PORTZ}]:50071\"
          - location: ${VETH_X}
            endpoint: \"[${OTG_PORTZ}]:5555;3+[${OTG_PORTZ}]:50071\"
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./config.yaml > /dev/null \
    && docker exec ixia-c-controller mkdir -p ${configdir} \
    && docker cp ./config.yaml ixia-c-controller:${configdir}/ \
    && rm -rf ./config.yaml
}

gen_config_common() {
    yml="otg_host: https://localhost
        otg_speed: speed_1_gbps
        otg_capture_check: true
        otg_iterations: 100
        otg_grpc_transport: false
        "
    echo -n "$yml" | sed "s/^        //g" | tee -a ./test-config.yaml > /dev/null
}

gen_config_b2b_dp() {
    yml="otg_ports:
          - ${VETH_A}
          - ${VETH_Z}
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null

    gen_config_common
}

gen_config_b2b_cpdp() {
    yml="otg_ports:
          - ${VETH_A}
          - ${VETH_Z}
        "
    echo -n "$yml" | sed "s/^        //g" | tee ./test-config.yaml > /dev/null

    gen_config_common
}

gen_config_b2b_lag() {
    yml="otg_ports:
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
    echo "Setting up back-to-back with CP/DP distribution of ixia-c ..."
    login_ghcr                                              \
    && docker run -d                                        \
        --name=ixia-c-controller                            \
        --publish 0.0.0.0:443:443                           \
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
    && gen_controller_config_b2b_cpdp                       \
    && gen_config_b2b_cpdp                                  \
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
        --publish 0.0.0.0:443:443                           \
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

ipv6_enable_docker() {
    echo "$(docker inspect bridge | grep "EnableIPv6")"
    echo "$(cat /etc/docker/daemon.json)"
    echo "{\"ipv6\": true, \"fixed-cidr-v6\": \"2001:db8:1::/64\"}" | sudo tee /etc/docker/daemon.json
    # echo "{\"ipv6\": false, \"fixed-cidr-v6\": \"2001:db8:1::/64\"}" > "/etc/docker/daemon.json"
    echo "$(cat /etc/docker/daemon.json)"
    echo "$(sudo systemctl restart docker)"
}

topo() {
    if [ "${3}" = "enable_ipv6" ]
    then 
        ipv6_enable_docker
        enable_ipv6=true
    fi
    case $1 in
        new )
            case $2 in
                dp  )
                    create_ixia_c_b2b_dp
                ;;
                cpdp)
                    create_ixia_c_b2b_cpdp
                ;;
                lag )
                    create_ixia_c_b2b_lag
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
                *   )
                    echo "unsupported topo type: ${2}"
                    exit 1
                ;;
            esac
        ;;
        logs    )
            mkdir -p logs/ixia-c-controller
            docker cp ixia-c-controller:/home/keysight/ixia-c/controller/logs/ logs/ixia-c-controller
            docker cp ixia-c-controller:/home/keysight/ixia-c/controller/config/config.yaml logs/ixia-c-controller
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
