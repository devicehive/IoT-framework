FROM alpine:3.2

# bootstrap
COPY ./init.sh /init.sh
RUN /init.sh v2

WORKDIR "/go/src/github.com/devicehive/IoT-framework/examples"
CMD ["/bin/sh"]
