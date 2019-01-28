FROM golang:1.11.1 AS Builder

RUN go get -u github.com/docker/docker/client

COPY . /go/src/app
WORKDIR /go/src/app

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build -ldflags '-w -extldflags "-static"'


FROM alpine:3.8 AS Runner

RUN apk add --update ca-certificates
COPY  --from=Builder /go/src/app/app /usr/local/bin/app

CMD [ "app" ]