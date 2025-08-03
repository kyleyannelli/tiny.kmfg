FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o tiny-kmfg .

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/tiny-kmfg .
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data /app/certs

EXPOSE 30108 30109

CMD ["./tiny-kmfg"]
