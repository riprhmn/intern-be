FROM golang:1.25-alpine AS builder
WORKDIR /src/services/magang

RUN apk add --no-cache git ca-certificates

# Cache dependencies separately. go.mod/go.sum must be tidied on the host and
# are never mutated during an image build.
COPY services/magang/go.mod services/magang/go.sum ./
ARG GOPROXY=direct
RUN GOPROXY=${GOPROXY} go mod download

COPY services/magang/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /bin/magang-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bin/magang-api /app/magang-api
COPY services/magang/config/config.yaml /app/config/config.yaml
EXPOSE 8090
CMD ["./magang-api"]
