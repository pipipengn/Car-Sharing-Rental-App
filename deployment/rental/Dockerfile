FROM golang:1.18-alpine AS builder

RUN go env -w GO111MODULE=on

COPY . /go/src/coolcar/server

WORKDIR /go/src/coolcar/server
RUN go install ./rental/...

FROM alpine:3.16
COPY --from=builder /go/bin/rental /bin/rental
ENV ADDR=:8081

EXPOSE 8081

ENTRYPOINT [ "/bin/rental" ]
