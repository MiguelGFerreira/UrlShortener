# Build stage: compiles one service selected via the SERVICE build arg.
FROM golang:1.23-alpine AS build

WORKDIR /src

# Download dependencies first to leverage layer caching.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE
RUN test -n "$SERVICE" || (echo "SERVICE build arg is required" && exit 1)
RUN CGO_ENABLED=0 go build -trimpath -o /app/service ./${SERVICE}

# Runtime stage: minimal image with just the static binary.
FROM alpine:3.20

RUN adduser -D -H appuser
USER appuser

COPY --from=build /app/service /app/service

ENTRYPOINT ["/app/service"]
