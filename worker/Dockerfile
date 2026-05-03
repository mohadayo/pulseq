FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o worker .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/worker .
EXPOSE 5001
CMD ["./worker"]
