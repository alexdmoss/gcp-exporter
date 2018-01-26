FROM golang:1.9

WORKDIR /tmp

RUN apt-get update && apt-get install -y git make bash wget build-essential
RUN wget -q https://get.docker.com/builds/Linux/x86_64/docker-17.05.0-ce.tgz -O /tmp/docker.tar.gz; \
    tar -xzf /tmp/docker.tar.gz -C /tmp/; \
    cp /tmp/docker/docker* /usr/bin; \
    chmod +x /usr/bin/docker*; \
    rm -rf /tmp/*
