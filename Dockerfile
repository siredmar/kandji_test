FROM 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/base-images:devpack-buster.latest
ENV GOOS=linux
ENV GOARCH=amd64

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

COPY ./bin/gxctl-${GOOS}-${GOARCH} /usr/bin/gxctl
RUN chmod +x /usr/bin/gxctl
RUN mkdir -p /root/.gxctl && \
  touch /root/.gxctl/config.yaml && \
  /usr/bin/gxctl ssh setup > /tmp/ssh_config && \
  sed -e s~"gxctl ssh tunnel --profile"~"gxctl ssh tunnel --skip-config-check -q --profile"~g /tmp/ssh_config >> /etc/ssh/ssh_config && \
  rm /root/.gxctl/config.yaml

ENTRYPOINT ["gxctl"]
