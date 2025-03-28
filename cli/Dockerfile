# Go build
FROM golang:1.24-bullseye AS gobuild

WORKDIR /app

COPY ./cli ./cli
COPY ./client ./client
COPY ./server ./server
COPY ./tools/preflight-checks.sh ./tools/preflight-checks.sh
COPY ./Makefile .

RUN make cli

# Copy binary to final image
FROM alpine:3.19

RUN apk add --no-cache tini bash

COPY --from=gobuild /app/build/admiral /usr/bin/admiral
COPY ./cli/scripts/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]
CMD [ "-h" ]
