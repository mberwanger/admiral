FROM alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c

ARG USER=admiral
ARG UID=1000
ARG GID=1000
ARG ADMIRAL_PORT=8080

ENV ADMIRAL_PORT=${ADMIRAL_PORT}

EXPOSE ${ADMIRAL_PORT}

HEALTHCHECK --interval=1m --timeout=3s --retries=3 \
  CMD curl -f http://localhost:${ADMIRAL_PORT}/healthcheck || exit 1

# Create non-root user and group
RUN addgroup -S -g "$GID" "$USER" && \
    adduser -S -u "$UID" -G "$USER" "$USER"

# Minimal upgrade & install only what's necessary
RUN apk add --no-cache \
      ca-certificates~=20241121-r1 \
      curl~=8 \
    && apk upgrade --no-cache \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

# Set working dir and switch to non-root user
WORKDIR /app
USER "$USER"

# Copy application files with ownership
COPY --chown=${USER}:${USER} admiral-server config.yaml /app/

# Set entrypoint and default command
ENTRYPOINT ["/app/admiral-server"]
CMD ["start", "--config", "config.yaml"]
