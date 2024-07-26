FROM golang:1.22.5-alpine

WORKDIR /mrbgpdv2
COPY . .
RUN go build -o ./mrbgpdv2
CMD ["./mrbgpdv2", "64513 10.200.100.3 65412 10.200.100.2 passive"]