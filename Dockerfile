FROM ubuntu:22.04

ENV LANG C

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y \
      git python3 python3-pip curl \
      g++ make autoconf pkg-config \
      libssl-dev tcl-dev libexpat1-dev \
      libpcre3-dev libtool libcap-dev graphviz \
      libluajit-5.1-dev libboost-dev libhwloc-dev default-libmysqlclient-dev \
      python3-distro libxml2-dev libncurses-dev libcurl4-openssl-dev libhiredis-dev \
      libkyotocabinet-dev libmemcached-dev libbrotli-dev \
      libcrypto++-dev libjansson-dev libcjose-dev libyaml-cpp-dev \
      libunwind-dev \
      python3-sphinx plantuml python3-sphinxcontrib.plantuml \
      libmaxminddb-dev \
      zlib1g-dev libgeoip-dev \
      gdb

ENV GO_VERSION 1.20.5
RUN curl -sSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar zxf - -C /usr/local/

ENV ATS_USER trafficserver
ENV ATS_HOME /home/trafficserver
RUN adduser --system --group --home $ATS_HOME $ATS_USER

ENV WORKSPACE $ATS_HOME/trafficserver

ARG GITHUB_REPO
ARG GITHUB_BRANCH

RUN git clone --depth 1 -b $GITHUB_BRANCH $GITHUB_REPO $WORKSPACE
WORKDIR $WORKSPACE
RUN git log -1 --pretty=fuller

WORKDIR $WORKSPACE
RUN autoreconf -if
# Some flags are copied from
# debian/rules in
# [trafficserver_9.1.1+ds-2build1.debian.tar.xz](http://archive.ubuntu.com/ubuntu/pool/universe/t/trafficserver/trafficserver_9.1.1+ds-2build1.debian.tar.xz)
# https://packages.ubuntu.com/jammy/trafficserver
RUN ./configure \
        --enable-layout=Debian \
        --sysconfdir=/etc/trafficserver --libdir=/usr/lib/trafficserver \
        --libexecdir=/usr/lib/trafficserver/modules \
        --with-user=root --with-group=root --disable-silent-rules \
        --enable-experimental-plugins --enable-32bit-build \
        --enable-mime-sanity-check \
        --with-hrw-geo-provider=maxminddb \
        --enable-werror --enable-debug --enable-wccp --enable-luajit
RUN make -j V=1

USER root
RUN make install
RUN chown trafficserver:trafficserver /run/trafficserver /var/cache/trafficserver /var/log/trafficserver

RUN mkdir $ATS_HOME/ats_go_test
COPY *.go go.mod go.sum $ATS_HOME/ats_go_test/
WORKDIR $ATS_HOME/ats_go_test

CMD ["/bin/bash"]
