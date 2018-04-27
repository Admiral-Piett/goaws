.DEFAULT_GOAL := run
VERSION=1.0.3
GITHUB_API_KEY = ${GITHUBB_API_KEY}
APIJSON='{"tag_name": "$(VERSION)","target_commitish": "master","name": "$(VERSION)","body": "Release of version  $(VERSION)","draft": true,"prerelease": true}'

dep:
	dep ensure

fmt:
	go fmt ./app/...

test:
	go test ./app/...

run: dep fmt test
	go run app/cmd/goaws.go

git-release:
	curl --data $(APIJSON) https://api.github.com/repos/p4tin/goaws/releases?access_token=$(GITHUB_API_KEY)

linux:
	GOOS=linux GOARCH=amd64 go build -o goaws_linux_amd64  app/cmd/goaws.go

docker-release: linux
	docker build -t pafortin/goaws .
	docker tag pafortin/goaws pafortin/goaws:$(VERSION)


