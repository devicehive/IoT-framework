FROM alpine:3.3

# bootstrap
COPY ./init.sh /init.sh
RUN /init.sh v2

# prepare scripts
COPY ./start.sh /start.sh

ENTRYPOINT ["/start.sh"]
