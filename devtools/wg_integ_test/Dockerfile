FROM linuxserver/wireguard:amd64-version-v1.0.20210424

WORKDIR /tmp
RUN curl -LO https://golang.org/dl/go1.16.5.linux-amd64.tar.gz \
  && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.16.5.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin:/root/go/bin
WORKDIR /
RUN touch /config/wg0.conf && chmod 600 /config/wg0.conf
LABEL org.opencontainers.image.source "https://github.com/soracom/soratun"

