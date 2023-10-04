FROM alpine:3.18

COPY ./build/rulestone_linux_amd64 /usr/local/bin/rulestone

RUN mkdir /rulestone
WORKDIR /rulestone

ENTRYPOINT ["/usr/local/bin/rulestone", "--rpc", "grpc", "-conn", "tcp"]

