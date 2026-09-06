FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN go -C services/magang mod download && go -C services/magang mod tidy
RUN go -C services/magang build -o /bin/magang-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bin/magang-api /app/magang-api
COPY services/magang/config/config.yaml /app/config/config.yaml
EXPOSE 8090
CMD ["./magang-api"]
