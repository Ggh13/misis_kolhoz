FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/internal/moked_data/farmers_sku.xlsx ./internal/moked_data/farmers_sku.xlsx
COPY --from=builder /app/internal/moked_data/orders.xlsx ./internal/moked_data/orders.xlsx

EXPOSE 8080
CMD ["./server"]