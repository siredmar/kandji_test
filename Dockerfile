FROM gridx/devpack:buster.latest-linux-amd64

# Install colordiff
RUN apt-get update && apt-get install -y --no-install-recommends colordiff && \
  rm -rf /var/lib/apt/lists/*

ENV YAML2JSON_VERSION=0.4.0
RUN wget https://github.com/wakeful/yaml2json/releases/download/${YAML2JSON_VERSION}/yaml2json-linux-amd64 && \
  mv yaml2json-linux-amd64 /usr/local/bin/yaml2json && \
  chmod +x /usr/local/bin/yaml2json

COPY ./bin/gxctl /usr/bin/gxctl
RUN chmod +x /usr/bin/gxctl

ENTRYPOINT ["gxctl"]
