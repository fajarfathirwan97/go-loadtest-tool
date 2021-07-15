.PHONY: dep run migrate rollback docs-update build

run:
	go run main.go

rollback:
	go run main.go rollback

migrate:
	go run main.go migrate

dep:
	go mod download
	go mod verify

docs-update:
	rm -rf docs./
	swag init

build:
	go build .
