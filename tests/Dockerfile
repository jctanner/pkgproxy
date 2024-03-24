FROM quay.io/centos/centos:stream9

COPY dnf.conf /tmp/dnf.conf
RUN cat /tmp/dnf.conf

RUN curl --proxy http://192.168.124.183:80 http:/tannerjc.net
RUN dnf --config=/tmp/dnf.conf -y install gettext
