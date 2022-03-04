FROM alpine AS runtime

FROM golang:alpine AS builder
COPY cmn /ez/cmn
COPY cmd /ez/cmd
COPY ezt /ez/ezt
COPY cadet /ez/cadet
COPY go.* /ez/
COPY Makefile /ez/
COPY swp /ez/swp
WORKDIR /ez
RUN go env -w CGO_ENABLED=0 && GOOS=linux GOARCH=amd64 \
    cd cmd/ez && go clean && go build && cd -; \
    cd cmd/ezl && go clean && go build && cd -; \
    cd cmd/ezt && go clean && go build && cd -; \
    cd cmd/ezs && go clean && go build

FROM runtime
COPY --from=builder /ez/cmd/ez/ez /usr/local/bin/ez
COPY --from=builder /ez/cmd/ezl/ezl /usr/local/bin/ezl
COPY --from=builder /ez/cmd/ezs/ezs /usr/local/bin/ezs
COPY --from=builder /ez/cmd/ezt/ezt /usr/local/bin/ezt
