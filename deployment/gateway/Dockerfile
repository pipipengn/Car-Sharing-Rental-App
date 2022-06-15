FROM golang:1.18-alpine AS builder

RUN go env -w GO111MODULE=on

COPY . /go/src/coolcar/server

WORKDIR /go/src/coolcar/server
RUN go install ./gateway...

FROM alpine:3.16
COPY --from=builder /go/bin/gateway /bin/gateway
ENV ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["/bin/gateway"]


