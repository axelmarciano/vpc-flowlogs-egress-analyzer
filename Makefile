DOCKER_FLAG := $(findstring docker, $(MAKECMDGOALS))
HTML_FLAG := $(findstring html, $(MAKECMDGOALS))
MAKEFLAGS += --silent

IMAGE_NAME := vpc-flowlogs-egress-analyzer

AWS_VOLUME := -v $(HOME)/.aws:/root/.aws:ro

build:
ifeq ($(DOCKER_FLAG),docker)
	docker build -t $(IMAGE_NAME) .
else
	go build ./...
endif

run:
ifeq ($(DOCKER_FLAG),docker)
	docker run --rm $(AWS_VOLUME) $(IMAGE_NAME)
else
	go run cmd/main.go
endif

watch:
ifeq ($(DOCKER_FLAG),docker)
	reflex -r '\.go$$' -s -- sh -c \
		"docker build -t $(IMAGE_NAME) . && docker run --rm $(AWS_VOLUME) $(IMAGE_NAME)"
else
	reflex -r '\.go$$' -s -- sh -c "go run cmd/main.go"
endif

