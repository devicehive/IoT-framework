FROM alpine:3.3

# bootstrap
COPY ./init.sh /init.sh
RUN /init.sh --branch=v2

WORKDIR "/go/src/github.com/devicehive/IoT-framework/examples"
ENV GOPATH=/go
CMD ["/bin/sh"]
