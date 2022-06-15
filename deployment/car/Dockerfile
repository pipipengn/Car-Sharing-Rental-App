FROM golang:1.18-alpine AS builder

RUN go env -w GO111MODULE=on

COPY . /go/src/coolcar/server

WORKDIR /go/src/coolcar/server
RUN go install ./car/...

FROM alpine:3.16
COPY --from=builder /go/bin/car /bin/car
ENV ADDR=:8081
ENV WS_ADDR=:8080

EXPOSE 8080
EXPOSE 8081

ENTRYPOINT [ "/bin/car" ]
