FROM golang AS baseforbuilder
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y protobuf-compiler && \
    go get google.golang.org/protobuf/cmd/protoc-gen-go

FROM baseforbuilder AS builder
COPY chunks /go-ez/chunks
COPY cli /go-ez/cli
COPY cmd /go-ez/cmd
COPY ezs /go-ez/ezs
COPY ezt /go-ez/ezt
COPY hash /go-ez/hash
COPY go.* /go-ez/
COPY Makefile /go-ez/
WORKDIR /go-ez
RUN GOOS=linux GOARCH=amd64 make clean && make

FROM debian:testing-slim
COPY --from=builder /go-ez/cmd/ez/ez /go-ez/bin/ez
COPY --from=builder /go-ez/cmd/ezl/ezl /go-ez/bin/cli/ezl
COPY --from=builder /go-ez/cmd/ezs/ezs /go-ez/bin/ezs
COPY --from=builder /go-ez/cmd/ezt/ezt /go-ez/bin/ezt
