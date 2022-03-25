#!/bin/bash -ex

cd $(dirname $0)

go mod tidy
go fmt ./...
