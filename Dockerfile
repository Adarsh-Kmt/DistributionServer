FROM golang:1.22.1-alpine

RUN apk update && \
    apk add --no-cache git

RUN mkdir /app

WORKDIR /app

COPY . /app

RUN go build -o binaryFile .

EXPOSE 9000

CMD ["/app/binaryFile"]