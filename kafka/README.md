# README
```bash
# docker compose up -d
# docker exec  -it 440b4cb8cb14 sh # 进入kafka生产者容器
$ kafka-topics --bootstrap-server localhost:9092 --create --topic test-topic # 创建topic
$ kafka-console-producer --bootstrap-server localhost:9092 --topic test-topic # 生产消息
>abc

# docker exec  -it 440b4cb8cb14 sh # 进入kafka消费者容器
$ kafka-console-consumer --bootstrap-server localhost:9092 --topic test-topic --from-beginning
abc
```