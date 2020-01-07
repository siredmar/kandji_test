FROM gridx/debian-base:buster.stable-linux-amd64

# Install colordiff
RUN apt-get update && apt-get install -y --no-install-recommends colordiff && \
  rm -rf /var/lib/apt/lists/*

COPY ./bin/gxctl /usr/bin/gxctl
RUN chmod +x /usr/bin/gxctl

ENTRYPOINT ["gxctl"]