#!/bin/sh

docker build --tag ngstaticserver . && docker run --publish 8080:8080 ngstaticserver