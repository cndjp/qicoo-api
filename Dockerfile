FROM alpine:3.8
MAINTAINER sugimount <https://twitter.com/sugimount>

RUN adduser -D qicoo-api
RUN apk --no-cache add tzdata
USER qicoo-api:qicoo-api
COPY ./bin/qicoo-api /home/qicoo-api/qicoo-api

EXPOSE 8080
ENTRYPOINT ["/home/qicoo-api/qicoo-api"]
