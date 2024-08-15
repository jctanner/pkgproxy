FROM golang:latest

RUN mkdir /src
RUN openssl genrsa -out /src/caKey.pem 2048
RUN openssl req -new -x509 -days 3650 -key /src/caKey.pem -out /src/caCert.pem -subj "/C=US/ST=New York/L=New York City/O=HaxxOrg/CN=haxx.net"
COPY . /src/
RUN cd /src; go build -o pkgproxyd ./cmd/pkgproxyd
WORKDIR /src
ENTRYPOINT ["./pkgproxyd"]
