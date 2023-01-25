.PHONY: deps
deps:
	go mod tidy

.PHONY: unit_test
unit_test:
	go test -count=1 -v ./... -parallel=4

.PHONY: lint
lint:
	curl https://gitlab.test.igdcs.com/finops/devops/cicd/runner/-/raw/master/.golangci.yml -o ./.golangci.yml
	golangci-lint run -v --config=./.golangci.yml
