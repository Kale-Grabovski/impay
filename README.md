## Wallets/Stats API

Run:

```bash
IMPAY_WALLETPORT=3421 IMPAY_KAFKA_HOST=localhost:9092 go run main.go wallet
IMPAY_STATSPORT=3422 IMPAY_KAFKA_HOST=localhost:9092 go run main.go stats
```

OR

```bash
docker compose up
```
