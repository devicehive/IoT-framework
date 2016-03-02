FROM alpine:3.3

# install D-Bus
RUN apk add --update dbus && rm -rf /var/cache/apk/*
RUN mkdir -p /var/run/dbus && \
	mkdir -p /etc/dbus-1/system.d

# prepare scripts
COPY ./start.sh /start.sh

VOLUME ["/etc/dbus-1/system.d", "/var/run/dbus"]

ENTRYPOINT ["/start.sh"]
CMD ["--system"]
