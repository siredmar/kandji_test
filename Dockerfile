FROM 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/base-images:devpack-buster.latest

# Install colordiff
RUN apt-get update && apt-get install -y --no-install-recommends colordiff && \
  rm -rf /var/lib/apt/lists/*

ENV YAML2JSON_VERSION=0.4.0
RUN wget https://github.com/wakeful/yaml2json/releases/download/${YAML2JSON_VERSION}/yaml2json-linux-amd64 && \
  mv yaml2json-linux-amd64 /usr/local/bin/yaml2json && \
  chmod +x /usr/local/bin/yaml2json

COPY ./bin/gxctl /usr/bin/gxctl
RUN chmod +x /usr/bin/gxctl
RUN mkdir -p /root/.gxctl && \
  touch /root/.gxctl/config.yaml && \
  /usr/bin/gxctl ssh setup > /tmp/ssh_config && \
  sed -e s~"gxctl ssh tunnel --profile"~"gxctl ssh tunnel -q --profile"~g /tmp/ssh_config >> /etc/ssh/ssh_config

ENTRYPOINT ["gxctl"]
