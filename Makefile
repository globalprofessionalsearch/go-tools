.PHONY: test

test:
	docker-compose run --rm dev go test ./...