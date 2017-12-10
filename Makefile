TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)
export TAG

test:
   go test ./...

