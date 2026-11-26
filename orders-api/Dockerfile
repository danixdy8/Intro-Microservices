FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o orders-api main.go

FROM alpine:3.21

WORKDIR /app
RUN mkdir -p /data
ENV DB_PATH=/data/orders.db
ENV PORT=8002
ENV GIN_MODE=release

COPY --from=builder /app/orders-api ./orders-api

EXPOSE 8002

CMD ["./orders-api"]
