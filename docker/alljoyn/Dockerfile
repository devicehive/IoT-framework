FROM alpine:3.3

# bootstrap
COPY ./init.sh /init.sh
COPY ./*.patch /tmp/
RUN /init.sh --branch=v2 --alljoyn=v15.04

# prepare scripts
COPY ./start.sh /start.sh

EXPOSE 9955 9956
ENTRYPOINT ["/start.sh"]
