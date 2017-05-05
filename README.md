# Chukonu

## Prerequisite
consul
docker
golang 1.8+

## Development
wget https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz
tar -xvfz go1.8.1.linux-amd64.tar.gz
sudo mv go /usr/local
mkdir -p gocode/src/github.com/xiaoyao1991/
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/gocode

consul agent -dev -client=0.0.0.0

docker run \
  --volume=/:/rootfs:ro \
  --volume=/var/run:/var/run:rw \
  --volume=/sys:/sys:ro \
  --volume=/var/lib/docker/:/var/lib/docker:ro \
  --publish=8080:8080 \
  --detach=true \
  --name=cadvisor \
  google/cadvisor:latest

./daemon -cadvisor http://<hostip>:8080/ -consul http://<hostip>:8500
