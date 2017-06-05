all: fmt build

build:
	go build .

gen:
	go generate

fmt:
	find . ! -path "*/vendor/*" -type f -name '*.go' -exec gofmt -l -s -w {} \;

clean:
	rm -vf xgopher

.PHONY: xgopher fmt clean build gen
