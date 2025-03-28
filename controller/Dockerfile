# Go build
FROM golang:1.24-bullseye AS gobuild

WORKDIR /app

COPY ./controller ./controller
COPY ./client ./client
COPY ./server ./server
COPY ./tools/preflight-checks.sh ./tools/preflight-checks.sh
COPY ./Makefile .

RUN make controller

# Copy binary to final image
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=gobuild /app/build/admiral-controller /app

CMD ["/app/admiral-controller"]