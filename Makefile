.PHONY: compile test install

compile:
	go build -o terraform-provider-readthedocs .

test:
	go test ./...

install:
	./scripts/build-provider.sh
