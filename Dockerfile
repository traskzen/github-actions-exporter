FROM golang:1.18 AS builder

WORKDIR /app
COPY . .
RUN bash ./build.sh

FROM alpine:latest AS release
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /app/bin/app .
CMD ["./app"]