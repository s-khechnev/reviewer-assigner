FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/app

FROM alpine:3.22

RUN apk add --no-cache bash make

WORKDIR /app

COPY --from=builder /app .

ADD https://github.com/pressly/goose/releases/download/v3.26.0/goose_linux_x86_64 /bin/goose
RUN chmod +x /bin/goose

RUN chmod +x ./entrypoint.sh

CMD ["./entrypoint.sh"]
