#!/bin/bash

# GET THE PKGPROXY CONTAINER IP
#PROXY_HOST=$(docker inspect pkgproxy | grep IPAddress | tail -n1 | awk -F\" '{print $4}')
PROXY_HOST="192.168.124.183"
export DNF_PROXY_URL="http://${PROXY_HOST}:80"
export PIP_PROXY_URL="https://${PROXY_HOST}:443"

envsubst < dnf.conf.template > dnf.conf

docker build \
    --progress=plain \
    --no-cache \
    -t buildtest .
