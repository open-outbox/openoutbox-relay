# brew install nats-server nats-cli

# nats stream add OUTBOX_STREAM \
#   --subjects "events.*" \
#   --storage file \
#   --retention limits \
#   --discard old \
#   --max-msgs -1 \
#   --max-bytes -1 \
#   --dupe-window 2m

#   nats stream view OUTBOX_STREAM

docker-compose exec nats-box nats stream add OUTBOX_EVENTS \
  --server nats:4222 \
  --subjects "event.>" \
  --storage file \
  --retention limits \
  --discard old \
  --dupe-window 2m \
  --timeout 10s
