FROM golang:1.15.5 as builder

RUN apt-get -y update && apt-get -y install upx


# Copy the Go Modules manifests


# Copy the go source
COPY go-alertmanager-action/  /app

WORKDIR /app

# Build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn"

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download && \
    go build -a -o alert_action main.go && \
    upx alert_action


FROM alpine:3.9.2

# 基础环境配置
RUN set -ex \
    # 修改系统时区为东八区
    && apk add -U tzdata \
    && rm -rf /etc/localtime \
    && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder /app/alert_action .

ENTRYPOINT ["/alert_action"]