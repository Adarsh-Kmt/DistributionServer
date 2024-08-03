FROM golang:1.22.1-alpine as builder


RUN apk update && \
    apk add --no-cache git

RUN mkdir /app

WORKDIR /app
COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go build -v -o binaryFile .



FROM alpine:latest as production
RUN apk --no-cache add ca-certificates

RUN mkdir /prod
WORKDIR /prod

COPY ./root-key.pem /prod/
COPY ./root.pem /prod/
COPY ./DistributionServer-key.pem /prod/
COPY ./DistributionServer.pem /prod/

COPY --from=builder /app/binaryFile /prod/

RUN cp /prod/root.pem /etc/ssl/certs/ && update-ca-certificates

ENV GRPC_GO_LOG_VERBOSITY_LEVEL=99
ENV GRPC_GO_LOG_SEVERITY_LEVEL=info

EXPOSE 9000

CMD ["/prod/binaryFile"]
