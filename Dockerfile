FROM debian:buster AS runtime
LABEL maintainer="Moov <support@moov.io>"

WORKDIR /

RUN apt-get update && apt-get install -y ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

COPY bin/.docker/customers /app/customers

EXPOSE 8087/tcp
EXPOSE 9097/tcp

VOLUME [ "/data", "/configs" ]

ENTRYPOINT ["/app/customers"]
