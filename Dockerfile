# Web build
FROM node:23-bullseye AS nodebuild

WORKDIR /app

COPY ./web ./web
COPY ./tools ./tools
COPY ./Makefile .

RUN make web

# Go build
FROM golang:1.24-bullseye AS gobuild

WORKDIR /app

COPY ./server ./server
COPY ./tools/preflight-checks.sh ./tools/preflight-checks.sh
COPY ./Makefile .

COPY --from=nodebuild /app/web/build ./web/build

RUN make server-with-assets

# Copy binary to final image
FROM gcr.io/distroless/base-debian12

EXPOSE 8080

WORKDIR /app

COPY --from=gobuild /app/build/server /app
COPY ./server/config.yaml /app

CMD ["/app/server"]

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1