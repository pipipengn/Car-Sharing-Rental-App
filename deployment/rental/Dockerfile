FROM golang:1.18-alpine AS builder

RUN go env -w GO111MODULE=on

COPY . /go/src/coolcar/server

WORKDIR /go/src/coolcar/server
RUN go install ./rental/...
RUN go install github.com/grpc-ecosystem/grpc-health-probe@latest

FROM alpine:3.16
COPY --from=builder /go/bin/rental /bin/rental
COPY --from=builder /go/bin/grpc-health-probe /bin/grpc-health-probe
ENV ADDR=:8081

EXPOSE 8081

ENTRYPOINT [ "/bin/rental" ]
