FROM golang:1.8.1

RUN apt-get -y update && apt-get -y install net-tools lsof

RUN mkdir -p /go/src/github.com/xiaoyao1991/chukonu

RUN go get golang.org/x/net/context && go get github.com/satori/go.uuid && go get github.com/google/cadvisor/client && go get github.com/hashicorp/consul && go get github.com/docker/docker/client

ADD . /go/src/github.com/xiaoyao1991/chukonu

WORKDIR /go/src/github.com/xiaoyao1991/chukonu
RUN go build lifecycle.go && go build -buildmode=plugin druidtestplan.go

ENTRYPOINT ["./lifecycle"]
