.PHONY: build compile clean run up down


# go coomands
compile:
	cd jals && go mod download && go build

clean:
	rm ./jals/jals || true

run:
	cd jals && go run main.go


# docker compose commands
build:
	docker compose build

up: build
	docker compose up -d

down:
	docker compose down