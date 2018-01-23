#!/usr/bin/env bash

docker exec -it ldp kafka-topics --create --zookeeper 192.168.99.100:2181 --topic zip --replication-factor 1 --partitions 3
docker exec -it ldp kafka-topics --create --zookeeper 192.168.99.100:2181 --topic zip_address --replication-factor 1 --partitions 1
docker exec -it ldp kafka-topics --list --zookeeper 192.168.99.100:2181
