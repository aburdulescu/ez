FROM golang AS builder
COPY cmn /ez/cmn
COPY cmd /ez/cmd
COPY ezt /ez/ezt
COPY cadet /ez/cadet
COPY go.* /ez/
COPY Makefile /ez/
COPY swp /ez/swp
WORKDIR /ez
RUN GOOS=linux GOARCH=amd64 make clean && make

FROM debian:testing-slim
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y \
    procps \
    iproute2 \
    iputils-ping \
    netcat
COPY --from=builder /ez/cmd/ez/ez /usr/local/bin/ez
COPY --from=builder /ez/cmd/ezl/ezl /usr/local/bin/ezl
COPY --from=builder /ez/cmd/ezs/ezs /usr/local/bin/ezs
COPY --from=builder /ez/cmd/ezt/ezt /usr/local/bin/ezt
