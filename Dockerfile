FROM public.ecr.aws/n0w6x8l1/base-images:golang-dev-1.17.latest AS config

ARG HOST_GOOS
ARG HOST_GOARCH
ARG SRC_BIN=bin/gxctl-${HOST_GOOS}-${HOST_GOARCH}

COPY ${SRC_BIN} /usr/bin/gxctl
RUN chmod +x /usr/bin/gxctl
RUN mkdir -p /root/.gxctl && \
  touch /root/.gxctl/config.yaml && \
  /usr/bin/gxctl ssh setup > /tmp/ssh_config && \
  sed -e s~"gxctl ssh tunnel --profile"~"gxctl ssh tunnel --skip-config-check -q --profile"~g /tmp/ssh_config >> /etc/ssh/ssh_config && \
  rm /root/.gxctl/config.yaml

FROM 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/base-images:devpack-buster.latest

ARG TARGET_GOOS
ARG TARGET_GOARCH
ARG SRC_BIN=bin/gxctl-${TARGET_GOOS}-${TARGET_GOARCH}

# Install colordiff and new OpenSSH from backports (OpenSSH from buster doesn't fully support
# the needed SSH config)
RUN echo "deb http://deb.debian.org/debian buster-backports main" >> /etc/apt/sources.list && \
      apt-get update && apt-get install -y --no-install-recommends colordiff && \
      apt-get install -y --no-install-recommends -t buster-backports openssh-client && \
  rm -rf /var/lib/apt/lists/*

ENV YAML2JSON_VERSION=0.4.0
RUN wget https://github.com/wakeful/yaml2json/releases/download/${YAML2JSON_VERSION}/yaml2json-linux-amd64 && \
  mv yaml2json-linux-amd64 /usr/local/bin/yaml2json && \
  chmod +x /usr/local/bin/yaml2json

COPY ${SRC_BIN} /usr/bin/gxctl

COPY --from=config /etc/ssh/ssh_config /etc/ssh/ssh_config

ENTRYPOINT ["gxctl"]
