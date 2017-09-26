FROM alpine:latest

MAINTAINER Edward Muller <edward@heroku.com>

WORKDIR "/opt"

ADD .docker_build/hhse /opt/bin/hhse

CMD ["/opt/bin/hhse"]

