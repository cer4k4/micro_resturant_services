FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /opt/delivery

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o deliveryService main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /opt/delivery

COPY --from=builder /opt/delivery .

EXPOSE 8083

CMD ["./deliveryService"]
