# 镜像依赖
FROM golang:1.19-alpine

# 作者
MAINTAINER "wallnut"

#
RUN mkdir /app
ADD . /app/
WORKDIR /app

#RUN apk add git
#RUN git sysconfig --global https.proxy http://127.0.0.1:8118
#RUN git sysconfig --global https.proxy https://127.0.0.1:8118
RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go build -o main ./main.go


CMD ["./main"]