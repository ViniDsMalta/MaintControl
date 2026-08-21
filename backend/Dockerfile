FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o maintcontrol-api ./api

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/maintcontrol-api /app/maintcontrol-api

EXPOSE 8080

CMD ["/app/maintcontrol-api"]
