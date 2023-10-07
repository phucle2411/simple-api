

# Build
FROM golang:1.21.2-alpine AS build

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o bin/simple-api

FROM alpine:3.14 AS deploy

WORKDIR /app

COPY --from=build /src/bin/ /app

CMD ["/app/simple-api"]
