VERSION:=$(shell cat VERSION)

LDFLAGS="-X main.appVersion=$(VERSION)"

all:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=$(LDFLAGS) -o simple-api --installsuffix cgo main.go
	docker build --platform=linux/amd64 -t zizou2411/simple-api:$(VERSION) .
