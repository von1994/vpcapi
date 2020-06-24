FROM golang:1.13-alpine

COPY . $GOPATH/src/github.com/onionpiece/vpcapi
WORKDIR $GOPATH/src/github.com/onionpiece/vpcapi/tests/server
RUN go mod vendor
RUN go build .

EXPOSE 8443
ENTRYPOINT ["./server"]
