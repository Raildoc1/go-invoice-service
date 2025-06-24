FROM --platform=linux/$TARGETARCH golang:1.24.0-alpine AS build-stage

ARG TARGETARCH
RUN echo $TARGETARCH

# common mod
WORKDIR /go-invoice-service/
COPY ./common/go.mod ./common/go.sum ./common/
WORKDIR /go-invoice-service/common
RUN go mod download

# service mod
WORKDIR /go-invoice-service/
COPY ./services/api-service/go.mod ./services/api-service/go.sum ./services/api-service/
WORKDIR /go-invoice-service/services/api-service/
RUN go mod download

# copy source code
WORKDIR /go-invoice-service/
COPY ./common/ ./common/
COPY ./services/api-service/ ./services/api-service/

# build
WORKDIR /go-invoice-service/services/api-service/
RUN go mod tidy

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build -o server ./cmd/main.go

FROM alpine:latest AS release-stage

WORKDIR /

COPY --from=build-stage /go-invoice-service/services/api-service/server ./server

ENV PORT_TO_LISTEN=8080
ENV HTTP_ADDRESS=":${PORT_TO_LISTEN}"
ENV STORAGE_ADDRESS="localhost:5000"
ENV JWT_PRIVATE_KEY="secret"
ENV PROMETHEUS_PORT=9090
ENV OTEL_COLLECTOR_ADDRESS="localhost:4318"

ENTRYPOINT ["./server"]

EXPOSE ${PORT_TO_LISTEN}
