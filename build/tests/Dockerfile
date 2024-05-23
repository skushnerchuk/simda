FROM golang:1.22

ENV CGO_ENABLED 1
ENV CONFIG_PATH /etc/simda/config.yml

COPY . /app
COPY ./configs/simda_linux.yml /etc/simda/config.yml
COPY ./build/tests/entrypoint.sh /

WORKDIR /app

RUN apt-get update && apt-get install libpcap-dev -y && \
    sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/bin && \
    go mod download && \
    task build-daemon && \
    mkdir -p /opt/simda && \
    cp -f /app/bin/simda /opt/simda/simda && \
    chmod +x /opt/simda/simda

CMD ["/entrypoint.sh"]
