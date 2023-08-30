FROM golang:1.21-bullseye

WORKDIR /

COPY . .
RUN go build -o switcher .


CMD ["./switcher"]
