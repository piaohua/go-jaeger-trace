#!/bin/bash

set -e

start() {
    #Jaeger UI can be accessed at http://localhost:16686.
    docker run -d --name jaeger \
        -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
        -p 5775:5775/udp \
        -p 6831:6831/udp \
        -p 6832:6832/udp \
        -p 5778:5778 \
        -p 16686:16686 \
        -p 14268:14268 \
        -p 9411:9411 \
        jaegertracing/all-in-one:1.13

    #Then navigate to http://localhost:8080.
    docker run --rm -it \
        --link jaeger \
        -p 8080-8083:8080-8083 \
        -e JAEGER_AGENT_HOST="jaeger" \
        -e JAEGER_AGENT_PORT=6831 \
        jaegertracing/example-hotrod:1.13 \
        all
}

stop() {
    CONTAINER=jaeger
    docker stop ${CONTAINER}
}

up() {
    docker-compose -f docker-compose.yml up -d
}

down() {
    docker-compose -f docker-compose.yml down
}

case $1 in
    build)
        build;;
    start)
        start;;
    stop)
        stop;;
    up)
        up;;
    down)
        down;;
    *)
        echo "./run.sh build|start|stop"
esac
