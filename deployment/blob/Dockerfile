FROM golang:1.18-alpine AS builder

RUN go env -w GO111MODULE=on

COPY . /go/src/coolcar/server

WORKDIR /go/src/coolcar/server
RUN go install ./blob/...

FROM alpine:3.16
COPY --from=builder /go/bin/blob /bin/blob
ENV ADDR=:8081

# 申明暴露的端口
EXPOSE 8081

# 设置服务入口
ENTRYPOINT [ "/bin/blob" ]
