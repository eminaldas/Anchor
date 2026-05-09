FROM golang:alpine AS builder

WORKDIR /app

# Bağımlılıkları kopyala ve indir
COPY go.mod go.sum ./
RUN go mod download

# Tüm kaynak kodunu kopyala ve derle
COPY . .
RUN go build -o server ./cmd/server/main.go

# Minimal bir imaj ile çalıştır
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
