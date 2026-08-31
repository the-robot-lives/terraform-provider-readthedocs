.PHONY: compile test

compile:
	go build -o terraform-provider-readthedocs .

test:
	go test ./...
