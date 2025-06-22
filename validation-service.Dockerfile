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
COPY ./services/validation-service/go.mod ./services/validation-service/go.sum ./services/validation-service/
WORKDIR /go-invoice-service/services/validation-service/
RUN go mod download

# copy source code
WORKDIR /go-invoice-service/
COPY ./common/ ./common/
COPY ./services/validation-service/ ./services/validation-service/

# build
WORKDIR /go-invoice-service/services/validation-service/
RUN go mod tidy

RUN CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build -o server -tags musl ./cmd/main.go

FROM alpine:latest AS release-stage

WORKDIR /

COPY --from=build-stage /go-invoice-service/services/validation-service/server ./server

ENV KAFKA_ADDRESS="localhost:9092"
ENV STORAGE_ADDRESS="localhost:9090"
ENV KAFKA_POLL_TIMEOUT_MS=100

ENTRYPOINT ["./server"]
