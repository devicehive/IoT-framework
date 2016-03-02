FROM alpine:3.3

# bootstrap
COPY ./init.sh /tmp/init.sh
RUN chmod +x /tmp/init.sh && \
	/tmp/init.sh v2 && \
	rm -rf /tmp/init.sh

# prepare scripts
COPY ./start.sh /start.sh
RUN chmod +x /start.sh

ENTRYPOINT ["/start.sh"]
