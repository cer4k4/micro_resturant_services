FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /order

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o orderService main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /order

COPY --from=builder /order .

EXPOSE 8082

CMD ["./orderService"]
