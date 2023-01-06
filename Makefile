SUBDIRS = $(notdir $(wildcard ./cmd/*))
IMAGE_NAME := $(addprefix gmalbrand/, $(notdir $(shell pwd)))
IMAGE_VERSION := $(shell git tag --sort=committerdate -l v* | tail -1)

.PHONY: all
all: clean dep build

.PHONY: build
build: $(addprefix ./bin/, $(SUBDIRS))

$(SUBDIRS): %:./bin/%

./bin/%: ./cmd/%/main.go
	go build -o $@ $^

.PHONY: dep
dep:
	go mod download

.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: docker-image
docker-image:
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) -t $(IMAGE_NAME):latest -f build/Dockerfile .

.PHONY: docker-push-multiarch
docker-push-multiarch:
	docker buildx build --push -t $(IMAGE_NAME):$(IMAGE_VERSION) -t $(IMAGE_NAME):latest --platform "linux/amd64,linux/arm64" -f build/Dockerfile .
