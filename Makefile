VET_FLAGS=-assign -atomic -bools -copylocks -errorsas -nilfunc -printf -stdmethods -unusedresult -unreachable -tests

all: build vet test

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ${VET_FLAGS} ./...

.PHONY: build vet test all