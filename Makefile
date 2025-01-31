IMAGE_REF := ghcr.io/dualstacks/gitea-config-wave:local
docker:
	DOCKER_BUILDKIT=1 docker build -t $(IMAGE_REF) -f Dockerfile.goreleaser .

up-base:
	docker compose up -d

down-base:
	docker compose down

up-test:
	docker compose -f docker-compose.test.yml up -d

down-test:
	docker compose -f docker-compose.test.yml down -v

build:
	go build -o gitea-config-wave .

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
	./gitea-config-wave pull DUALSTACKS/reference-repo --config gitea-config-wave.yaml

push:
	./gitea-config-wave push --config gitea-config-wave.yaml

release-test:
	goreleaser release --snapshot --clean
