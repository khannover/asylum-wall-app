FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod ./
COPY main.go ./
COPY internal/ ./internal/
# web assets are embedded via internal/web/embed.go

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o asylum-wall .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates git openssh-client wget \
    && addgroup -S asylum && adduser -S asylum -G asylum

WORKDIR /app

COPY --from=builder /build/asylum-wall /app/asylum-wall

RUN mkdir -p /repo && chown -R asylum:asylum /repo /app

USER asylum

ENV REPO_PATH=/repo
ENV PORT=8080

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/app/asylum-wall"]