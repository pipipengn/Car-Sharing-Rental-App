FROM golang:1.18-alpine AS builder

RUN go env -w GO111MODULE=on

COPY . /go/src/coolcar/server

WORKDIR /go/src/coolcar/server
RUN go install ./auth/...


FROM alpine:3.16
COPY --from=builder /go/bin/auth /bin/auth
ENV ADDR=:8081

EXPOSE 8081

ENTRYPOINT [ "/bin/auth" ]
