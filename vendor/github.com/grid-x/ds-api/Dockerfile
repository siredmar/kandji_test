FROM gridx/debian-base:stretch.stable-linux-amd64

COPY bin/server /usr/local/bin/server
RUN chmod +x /usr/local/bin/server

ENTRYPOINT ["/usr/local/bin/server"]