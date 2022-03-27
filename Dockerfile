# build stage
FROM golang:1.17-alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /go/bin/app

# final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
ENTRYPOINT /app
LABEL Name=annoyer2server
EXPOSE 8080