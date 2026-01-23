FROM ubuntu:24.04

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /deploy

COPY Makefile main.tf ./
