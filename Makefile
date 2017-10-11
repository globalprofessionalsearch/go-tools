.PHONY: test

test:
	docker-compose run --rm dev go test ./...

deps:
	docker-compose run --rm dev dep ensure

deps-update:
	docker-compose run --rm dev dep ensure -update
