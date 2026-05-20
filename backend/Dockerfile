FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/knowledge-api ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -H appuser

COPY --from=build /bin/knowledge-api /app/knowledge-api

ENV PORT=8080
EXPOSE 8080

USER appuser

CMD ["/app/knowledge-api"]
