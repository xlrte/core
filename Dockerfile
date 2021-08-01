FROM golang:1.16.5-alpine3.13 AS builder

RUN apk update && apk add git
WORKDIR /build
COPY . .
RUN go build cmd/main.go

FROM alpine:3.13

RUN adduser -S hello

RUN apk update
WORKDIR /app

COPY --from=builder /build/main /app/hello

USER hello

CMD ["./hello"]
