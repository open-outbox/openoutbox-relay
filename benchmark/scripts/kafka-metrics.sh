docker compose exec kafka bash -c "
unset KAFKA_JMX_OPTS JMX_PORT KAFKA_JMX_PORT
echo 'Time     | EPS (1m Avg) | Throughput (MB/s)'
echo '------------------------------------------'
/opt/kafka/bin/kafka-run-class.sh kafka.tools.JmxTool \
  --object-name 'kafka.server:type=BrokerTopicMetrics,name=MessagesInPerSec,topic=openoutbox.events.v1' \
  --object-name 'kafka.server:type=BrokerTopicMetrics,name=BytesInPerSec,topic=openoutbox.events.v1' \
  --attributes OneMinuteRate \
  --reporting-interval 1000 \
  --jmx-url service:jmx:rmi:///jndi/rmi://127.0.0.1:9101/jmxrmi | awk -F',' '/^[0-9]/ {
    # Based on your raw output:
    # \$1 = Epoch Timestamp
    # \$2 = BytesInPerSec
    # \$3 = MessagesInPerSec

    byte_rate = \$2;
    msg_rate = \$3;
    mb_rate = byte_rate / 1024 / 1024;

    printf \"%s | %10.2f EPS | %10.2f MB/s\\n\", strftime(\"%H:%M:%S\"), msg_rate, mb_rate
    fflush()
  }'"
