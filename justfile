go := env_var_or_default('GOCMD', 'go')

default: tidy test

tidy:
	{{go}} mod tidy
	goimports -l -w {{justfile_directory()}}
	gofumpt -l -w {{justfile_directory()}}
	{{go}} fmt {{justfile_directory()}}/...

test:
	{{go}} vet {{justfile_directory()}}/...
	golangci-lint run {{justfile_directory()}}/...
	{{go}} test -race -coverprofile={{justfile_directory()}}/cover.out -timeout=7s -parallel=10 {{justfile_directory()}}/...
	{{go}} tool cover -html={{justfile_directory()}}/cover.out -o={{justfile_directory()}}/cover.html

todo:
	git -C {{justfile_directory()}} grep -e TODO --and --not -e ignoretodo
