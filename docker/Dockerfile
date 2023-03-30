FROM golang:1.20-alpine AS build

RUN apk add git
RUN apk --update --no-cache add g++

RUN mkdir /scratch
WORKDIR /scratch

RUN mkdir /app

COPY . .

RUN go mod download
RUN go mod verify

RUN go build -o /app/evm-transfer-logger

# Build image
FROM alpine:latest

EXPOSE 8080/tcp

COPY --from=build /app /app

WORKDIR /app

ENTRYPOINT [ "/app/evm-transfer-logger" ]