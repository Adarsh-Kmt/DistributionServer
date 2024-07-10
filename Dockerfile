FROM golang:1.22.1-alpine

RUN apk update && \
    apk add --no-cache git

RUN mkdir /app

WORKDIR /app

COPY . /app

RUN go build -o binaryFile .
ENV GRPC_GO_LOG_VERBOSITY_LEVEL=99
ENV GRPC_GO_LOG_SEVERITY_LEVEL=info
EXPOSE 9000

RUN cp /app/root.pem /etc/ssl/certs/ && update-ca-certificates

CMD ["/app/binaryFile"]