FROM golang:1.25.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o db-backup main.go

FROM alpine:3.18

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/db-backup .
COPY .env . 

ENV TZ=Asia/Jakarta

RUN chmod +x db-backup

ENTRYPOINT ["./db-backup", "backup"]