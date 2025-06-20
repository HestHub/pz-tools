FROM golang:1.24 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /pz-discord-bot .

# Todo test with
# FROM scratch
FROM golang:1.24 

COPY --from=builder /pz-discord-bot /pz-discord-bot

CMD ["/pz-discord-bot"]
