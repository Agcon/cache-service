FROM ubuntu:latest
LABEL authors="agcon"

ENTRYPOINT ["top", "-b"]