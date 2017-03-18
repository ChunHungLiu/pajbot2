PACKAGES = $(shell find ./ -type d -not -path '*/\.*')

build:
	glide install
	glide update
	go build

install:
	go install

test:
	go test -v $(shell go list ./... | grep -v /vendor/)

cover:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out
