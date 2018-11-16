.PHONY: test

test:
	docker-compose run --rm dev go test ./...

test-jsonio:
	docker-compose run --rm dev go test ./http/jsonio

deps:
	docker-compose run --rm dev dep ensure

deps-update:
	docker-compose run --rm dev dep ensure -update
