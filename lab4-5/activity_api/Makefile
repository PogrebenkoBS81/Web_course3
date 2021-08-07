.PHONY: local-up

local-up:
	@echo ">> Starting containers locally..."
	docker-compose up

.PHONY: test

test:
	echo ">> Running tests..."
	go test ./... -v
