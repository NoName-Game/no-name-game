FROM arm64v8/golang:1.15-alpine AS builder

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
ENV PROJECT_PATH=/go/src/nn-telegram

RUN mkdir -p $GOPATH/src/nn-grpc && \
    mkdir -p $PROJECT_PATH

COPY . $PROJECT_PATH
COPY ./nn-grpc $GOPATH/src/nn-grpc

# go build server
WORKDIR $PROJECT_PATH/cmd/client
RUN GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags=nn-telegram-client -i -installsuffix cgo -o /go/bin/client .

# go build enemy_respawner
WORKDIR $PROJECT_PATH/cmd/notifier
RUN GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags=nn-telegram-notifier -i -installsuffix cgo -o /go/bin/notifier .


FROM arm64v8/alpine:latest

RUN apk add --update coreutils tzdata && rm -rf /var/cache/apk/*

WORKDIR /

COPY --from=builder /go/src/nn-telegram/resources resources

COPY --from=builder /go/bin/client /client
COPY --from=builder /go/bin/notifier /notifier

STOPSIGNAL SIGTERM