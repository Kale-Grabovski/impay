FROM golang:1.21.0 as base

WORKDIR /tmp/impay

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o impay .

FROM debian:bookworm-slim

COPY --from=base /tmp/impay/impay .
COPY --from=base /tmp/impay/config-example.yaml .

CMD ["./impay", "wallet"]
