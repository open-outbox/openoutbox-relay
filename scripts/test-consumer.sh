# docker compose exec nats-box nats sub -s nats:4222 ">"
docker-compose exec nats-box nats -s nats:4222 stream view OUTBOX_EVENTS
