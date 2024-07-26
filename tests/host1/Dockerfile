FROM golang:1.22.5-alpine

WORKDIR /mrbgpdv2
COPY ../cmd/. .
RUN go build -o ./mrbgpdv2
CMD ["./mrbgpdv2", "64512 10.200.100.2 65413 10.200.100.3 active"]