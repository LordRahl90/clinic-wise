FROM golang:alpine AS build

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

RUN CGO_ENABLE=0 go build -o api ./cmd/api
RUN CGO_ENABLE=0 go build -o migration ./cmd/migrations


FROM gcr.io/distroless/base-debian10

WORKDIR /

# Copy certificates
COPY --from=build /etc/ssl/certs /etc/ssl/certs

COPY --from=build /app/api api
COPY --from=build /app/migration migration

EXPOSE 8080