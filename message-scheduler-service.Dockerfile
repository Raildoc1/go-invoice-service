FROM --platform=linux/$TARGETARCH golang:1.24.0-alpine AS build-stage

ARG TARGETARCH
RUN echo $TARGETARCH

RUN apk update
RUN apk add \
    gcc \
    musl-dev

# common mod
WORKDIR /go-invoice-service/
COPY ./common/go.mod ./common/go.sum ./common/
WORKDIR /go-invoice-service/common
RUN go mod download

# service mod
WORKDIR /go-invoice-service/
COPY ./services/message-scheduler-service/go.mod ./services/message-scheduler-service/go.sum ./services/message-scheduler-service/
WORKDIR /go-invoice-service/services/message-scheduler-service/
RUN go mod download

# copy source code
WORKDIR /go-invoice-service/
COPY ./common/ ./common/
COPY ./services/message-scheduler-service/ ./services/message-scheduler-service/

# build
WORKDIR /go-invoice-service/services/message-scheduler-service/
RUN go mod tidy

RUN CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build -o server -tags musl ./cmd/main.go

FROM alpine:latest AS release-stage

WORKDIR /

COPY --from=build-stage /go-invoice-service/services/message-scheduler-service/server ./server

ENV KAFKA_ADDRESS="localhost:9092"
ENV STORAGE_ADDRESS="localhost:9090"
ENV WORKERS_COUNT=3
ENV RETRY_INTERVAL_MS=30000
ENV DISPATCH_INTERVAL_MS=1000

ENTRYPOINT ["./server"]
