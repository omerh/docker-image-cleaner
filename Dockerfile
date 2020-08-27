FROM golang:1.15 AS Builder

ENV GO111MODULE=on

COPY . /go/src/app
WORKDIR /go/src/app

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build -ldflags '-w -extldflags "-static"' -o app

FROM scratch

COPY  --from=Builder /go/src/app/app /go/bin/app

CMD [ "/go/bin/app" ]
