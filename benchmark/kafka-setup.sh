docker exec -it benchmark-kafka-1 /opt/kafka/bin/kafka-topics.sh \
  --create \
  --topic outbox.events.v1 \
  --bootstrap-server localhost:9092 \
  --partitions 12 \
  --replication-factor 1