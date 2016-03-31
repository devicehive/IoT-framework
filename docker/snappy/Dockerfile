FROM ubuntu:15.10

# install snappy tools and snapcraft
RUN apt-get install -y software-properties-common
RUN apt-add-repository ppa:snappy-dev/tools && apt update
RUN apt-get install -y snappy-tools snapcraft

# development tools
RUN apt-get install -y vim golang-go git

ENV GOPATH /go
ENV BRANCH v2

RUN mkdir -p $GOPATH/src/github.com/devicehive
WORKDIR $GOPATH/src/github.com/devicehive
RUN git clone -b $BRANCH --single-branch https://github.com/devicehive/IoT-framework

WORKDIR $GOPATH/src/github.com/devicehive/IoT-framework
RUN go get -d -t -v ./...

ENTRYPOINT ["/bin/bash"]
