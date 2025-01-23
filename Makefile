.PHONY: all build clean test docker-build docker-run

APP_NAME := docker-auto-backup
DOCKER_IMAGE := bytekai/docker-auto-backup
TAG := latest

all: build

build:
	go build -o $(APP_NAME)

clean:
	rm -f $(APP_NAME)
	docker rm -f $(APP_NAME) || true
	docker rmi -f $(DOCKER_IMAGE):$(TAG) || true


test:
	go test -v ./...

docker-build:
	docker build -t $(DOCKER_IMAGE):$(TAG) . --progress=plain

docker-run: docker-build
	docker rm -f $(APP_NAME) || true
	docker run \
		--name $(APP_NAME) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PWD)/backups:/backups \
		$(DOCKER_IMAGE):$(TAG)

dev: build
	./$(APP_NAME)
