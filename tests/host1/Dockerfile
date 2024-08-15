FROM golang:1.22.5-alpine

WORKDIR /gobgp
COPY ../cmd/. .
RUN go mod tidy
RUN go build -o ./gobgp
CMD ["./gobgp", "64512 10.200.100.2 65413 10.200.100.3 active"]