FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o spiffe-client .

FROM alpine:latest

WORKDIR /root

RUN apk add --no-cache bash

COPY --from=builder /app/spiffe-client .

EXPOSE 8080

CMD ["./spiffe-client"]
