IMAGE_NAME := render
REPOSITORY := quay.io/sergey_grebenshchikov/$(IMAGE_NAME)

VERSION = latest
TAG ?= $(VERSION)

PACKAGES := $(shell go list -f {{.Dir}} ./...)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

.PHONY: all clean build push

all: build

push:
	docker push $(REPOSITORY):$(TAG)
	
build: bin/linux_amd64/render
	docker build -t $(IMAGE_NAME):$(TAG) -t $(REPOSITORY):$(TAG) .
	
bin/render: $(GOFILES)
	go build -ldflags "-X main.version=$(VERSION)"  -o bin/render ./cmd/render

bin/linux_amd64/render: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)"  -o bin/linux_amd64/render ./cmd/render

clean: 
	rm -rf bin/*
