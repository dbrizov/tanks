# syntax=docker/dockerfile:1

# --- Build stage: compile a static binary ---
FROM golang:1.26 AS build
WORKDIR /src

# Cache module downloads separately from the source for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO off -> a fully static binary that runs on the distroless "static" base.
# Built natively on the host arch (arm64 on the Oracle VM); no cross-compile.
RUN CGO_ENABLED=0 go build -o /out/server ./...

# --- Runtime stage: minimal, no shell, non-root ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
EXPOSE 8080
# JWT_SECRET is supplied at runtime via env (see server/auth.go).
ENTRYPOINT ["/server"]
