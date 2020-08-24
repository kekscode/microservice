#!/bin/sh

go build -o guzzler *.go
GUZZLER_NAME=guz_1 GUZZLER_BIND="localhost:8080" GUZZLER_PEERS="http://localhost:8081,http://localhost:8082" ./guzzler &
GUZZLER_NAME=guz_2 GUZZLER_BIND="localhost:8081" GUZZLER_PEERS="http://localhost:8080,http://localhost:8082" ./guzzler &
GUZZLER_NAME=guz_3 GUZZLER_BIND="localhost:8082" GUZZLER_PEERS="http://localhost:8080,http://localhost:8081" ./guzzler &
