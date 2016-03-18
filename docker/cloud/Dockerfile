FROM alpine:3.3

# bootstrap
COPY ./init.sh /init.sh
RUN /init.sh --branch=v2

# prepare config and scripts
COPY ./config.yml /etc/devicehive-cloud.yml
COPY ./start.sh /start.sh

ENTRYPOINT ["/start.sh"]
CMD ["--conf", "/etc/devicehive-cloud.yml"]
