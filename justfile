go := env_var_or_default('GOCMD', 'go')

default: tidy test

tidy:
	{{go}} mod tidy
	goimports -l -w {{justfile_directory()}}
	gofumpt -l -w {{justfile_directory()}}
	{{go}} fmt {{justfile_directory()}}/...

test:
	{{go}} vet {{justfile_directory()}}/...
	# TODO: Re-enable golangci-lint once it's compatible with go 1.20
	# golangci-lint run {{justfile_directory()}}/...
	{{go}} test -race -timeout=7s -parallel=10 {{justfile_directory()}}/...

todo:
	git -C {{justfile_directory()}} grep -e TODO --and --not -e ignoretodo
