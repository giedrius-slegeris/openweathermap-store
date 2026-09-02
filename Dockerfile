# syntax=docker/dockerfile:1

# ---------- build stage ----------
FROM golang:1.27-alpine3.24 AS build

WORKDIR /src

# Resolve dependencies first so this layer stays cached until go.mod/go.sum change
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# CGO_ENABLED=0 produces a static binary, so the runtime image needs no libc shim
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /openweathermap-store .

# ---------- runtime stage ----------
FROM alpine:3.24

# ca-certificates: HTTPS calls to the OpenWeatherMap API
# tzdata: time.LoadLocation(TIMEZONE) in the cron scheduler
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /openweathermap-store /usr/local/bin/openweathermap-store

USER nobody

ENTRYPOINT ["openweathermap-store"]
