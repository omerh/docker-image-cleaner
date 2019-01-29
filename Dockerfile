FROM golang:1.11.1 AS Builder

RUN go get -u github.com/docker/docker/client

COPY . /go/src/app
WORKDIR /go/src/app

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build -ldflags '-w -extldflags "-static"'


FROM scratch

COPY  --from=Builder /go/src/app/app /go/bin/app

CMD [ "/go/bin/app" ]