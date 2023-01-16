FROM 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/base-images:devpack-buster.latest

ARG SRC_BIN
ARG SRC_CONFIG=bin/ssh_config

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

COPY ${SRC_CONFIG} /etc/ssh/ssh_config
COPY ${SRC_BIN} /usr/bin/gxctl
RUN chmod +x /usr/bin/gxctl

ENTRYPOINT ["gxctl"]
