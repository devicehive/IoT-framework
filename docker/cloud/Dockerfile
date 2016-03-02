FROM alpine:3.3

# bootstrap
COPY ./init.sh /tmp/init.sh
RUN chmod +x /tmp/init.sh && \
	/tmp/init.sh v2 && \
	rm -rf /tmp/init.sh

# prepare scripts and config
COPY ./start.sh /start.sh
RUN chmod +x /start.sh
COPY ./config.yml /etc/devicehive-cloud.yml

ENTRYPOINT ["/start.sh"]
# CMD ["--conf", "/etc/devicehive-cloud.yml"]
