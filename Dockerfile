#FROM ubuntu:latest
#LABEL authors="max_m"
#
#ENTRYPOINT ["top", "-b"]

# build stage

#
#FROM golang:1.26-alpine AS builder
#
#WORKDIR /app
#
#COPY go.mod ./
#RUN go mod download
#
#COPY . .
#
#RUN go build -o app ./cmd/auth/main.go
#
#
## run stage
#FROM alpine:3.20
#
#WORKDIR /app
#
#COPY --from=builder /app/app .
#
#EXPOSE 8081
#
#CMD ["./auth-service"]


# ================== BUILD STAGE ==================
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN rm -rf migrations

# Оптимизированная сборка
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /app/app ./cmd/auth/main.go

# ================== FINAL STAGE ==================
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -H -u 1000 appuser

WORKDIR /app

COPY --from=builder /app/app .

RUN chown -R appuser:appuser /app && chmod +x /app/app

USER appuser

EXPOSE 8080

CMD ["./app"]