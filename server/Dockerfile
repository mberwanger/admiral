# Web build
FROM oven/bun:latest AS webbuild

WORKDIR /app

COPY ./web ./web
COPY ./tools ./tools
COPY ./Makefile .

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      make \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/

RUN make web

# Go build
FROM golang:1.24-bullseye AS gobuild

WORKDIR /app

COPY ./server ./server
COPY ./tools/preflight-checks.sh ./tools/preflight-checks.sh
COPY ./Makefile .

COPY --from=webbuild /app/web/build ./web/build

RUN make server-with-assets

# Copy binary to final image
FROM gcr.io/distroless/base-debian12

EXPOSE 8080

WORKDIR /app

COPY --from=gobuild /app/build/admiral-server /app
COPY ./server/config.yaml /app

CMD ["/app/admiral-server"]

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1