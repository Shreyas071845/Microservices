FROM golang:alpine AS builder
WORKDIR /workspace
COPY go.work go.work.sum ./
COPY api_gateway/ api_gateway/
COPY services/ services/
COPY pkg/ pkg/
WORKDIR /workspace/services/tracking_service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /grpc-server ./cmd/server

FROM alpine:latest
WORKDIR /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
COPY --from=builder /grpc-server ./grpc-server
CMD ["./grpc-server"]
