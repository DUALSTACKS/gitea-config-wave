IMAGE_REF := ghcr.io/dualstacks/gitea-config-wave:local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

docker:
	DOCKER_BUILDKIT=1 docker build -t $(IMAGE_REF) -f Dockerfile.goreleaser .

up-base:
	docker compose up -d

down-base:
	docker compose down

destroy-base:
	docker compose down -v

build:
	go build -ldflags "-X github.com/DUALSTACKS/gitea-config-wave/cmd.Version=$(VERSION)" -o gitea-config-wave .

version:
	./gitea-config-wave --version

format:
	go fmt ./...

vet:
	go vet ./...

test: build up-test
	go test -v ./test/integration/...
	$(MAKE) down-test

x:
	repomix

pull:
	./gitea-config-wave pull DUALSTACKS/.gitea --config gitea-config-wave.yaml

push:
	./gitea-config-wave push --config gitea-config-wave.yaml

release-test:
	goreleaser release --snapshot --clean
