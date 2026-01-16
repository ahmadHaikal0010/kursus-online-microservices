FROM golang:1.22-alpine

WORKDIR /app

# 1. Copy go.mod + go.sum
COPY go.mod go.sum ./

# 2. Download dependency
RUN go mod download

# 3. Copy source code
COPY . .

# 4. Build binary
RUN go build -o app

EXPOSE 8080
CMD ["./app"]
