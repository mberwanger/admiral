FROM alpine:3.21

ARG USER=admiral
ARG UID=1000
ARG GID=1000

# Create non-root user and group
RUN addgroup -S -g "$GID" "$USER" && \
    adduser -S -u "$UID" -G "$USER" "$USER"

# Minimal upgrade & install only what's necessary
RUN apk --no-cache add --upgrade ca-certificates && \
    update-ca-certificates

# Set working dir and switch to non-root user
WORKDIR /app
USER "$USER"

# Copy statically-linked binary
COPY --chown=$USER:$USER admiral-server /app
COPY --chown=$USER:$USER config.yaml /app

CMD ["/app/admiral-server", "start", "--config", "config.yaml"]

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1