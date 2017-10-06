FROM golang:1.9-alpine

# need git installed in order to fetch/manage deps w/ `go get`, or the new dep tool
RUN apk update && apk add --no-cache \
  # add any other deps here
  git

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/globalprofessionalsearch/go-tools
