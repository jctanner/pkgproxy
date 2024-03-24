#!/bin/bash

if [[ ! -f caKey.pem ]]; then
    openssl genrsa -out caKey.pem 2048
fi

if [[ ! -f caCert.pem ]]; then
    openssl req -new -x509 -days 3650 -key caKey.pem -out caCert.pem -subj "/C=US/ST=New York/L=New York City/O=HaxxOrg/CN=haxx.net"
fi

go run ./cmd/pkgproxyd
