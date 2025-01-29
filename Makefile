up-base:
	docker compose up -d

down-base:
	docker compose down

build:
	go build -o gitea-config-wave .

format:
	go fmt ./...

x:
	repomix

pull:
	./gitea-config-wave pull DUALSTACKS/test-gitea-config-wave --config gitea-config-wave.yaml

push:
	./gitea-config-wave push --config gitea-config-wave.yaml
