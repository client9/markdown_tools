
build: hooks  ## build, install, lint
	go install ./cmd/md
	gometalinter \
                 --vendor \
                 --deadline=60s \
                 --disable-all \
                 --enable=vet \
                 --enable=golint \
                 --enable=gofmt \
                 --enable=goimports \
                 --enable=gosimple \
                 --enable=staticcheck \
                 --enable=ineffassign \
                 ./...
	go test .


test:  ## run all tests
	go test .

clean:  ## clean up time
	rm -rf dist/ bin/
	go clean ./...
	git gc --aggressive

.git/hooks/pre-commit: scripts/pre-commit.sh
	cp -f scripts/pre-commit.sh .git/hooks/pre-commit
.git/hooks/commit-msg: scripts/commit-msg.sh
	cp -f scripts/commit-msg.sh .git/hooks/commit-msg
hooks: .git/hooks/pre-commit .git/hooks/commit-msg  ## install git precommit hooks

.PHONY: help ci console docker-build bench

# https://www.client9.com/self-documenting-makefiles/
help:
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
.DEFAULT_GOAL=help
.PHONY=help

