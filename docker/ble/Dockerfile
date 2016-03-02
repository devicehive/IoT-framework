FROM alpine:3.3

# bootstrap
COPY ./init.sh /init.sh
RUN chmod +x /init.sh && \
	/init.sh v2 && \
	rm -rf /init.sh

# prepare scripts
COPY ./start.sh /start.sh
RUN chmod +x /start.sh

ENTRYPOINT ["/start.sh"]
