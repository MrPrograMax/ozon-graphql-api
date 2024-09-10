FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./server.go

FROM scratch

COPY --from=builder /app/server .

COPY --from=builder /app/configs /configs
COPY --from=builder /app/.env /app/.env


EXPOSE 8080

CMD ["./server"]
