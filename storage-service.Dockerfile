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
COPY ./services/storage-service/go.mod ./services/storage-service/go.sum ./services/storage-service/
WORKDIR /go-invoice-service/services/storage-service/
RUN go mod download

# copy source code
WORKDIR /go-invoice-service/
COPY ./common/ ./common/
COPY ./services/storage-service/ ./services/storage-service/

# build
WORKDIR /go-invoice-service/services/storage-service/
RUN go mod tidy

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build -o server ./cmd/main.go

FROM alpine:latest AS release-stage

WORKDIR /

COPY --from=build-stage /go-invoice-service/services/storage-service/server ./server

ENTRYPOINT ["./server"]
