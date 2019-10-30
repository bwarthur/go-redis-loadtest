FROM golang:1.12

ENV GO111MODULE=on

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["redis-load"]
