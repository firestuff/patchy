default: tidy test

tidy:
	go mod tidy
	go fmt {{justfile_directory()}}/...

test:
	go vet {{justfile_directory()}}/...
	go test -race -timeout=7s -parallel=10 {{justfile_directory()}}/...