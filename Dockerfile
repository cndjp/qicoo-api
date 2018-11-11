FROM alpine:3.8
MAINTAINER sugimount <https://twitter.com/sugimount>

RUN adduser --uid 1000 -D qicoo-api
RUN apk --no-cache add tzdata
USER 1000:1000
COPY ./bin/qicoo-api /home/qicoo-api/qicoo-api

EXPOSE 8080
ENTRYPOINT ["/home/qicoo-api/qicoo-api"]
