#!/bin/bash
set -euC
set -o xtrace

for d in $(go list ./...); do
    go test -race -coverprofile=profile.out -covermode=atomic "$d"
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

# Only continue if we're on go 1.9; no need to run the linter for every case
if go version | grep -q go1.9; then
    diff -u <(echo -n) <(gofmt -e -d $(find . -type f -name '*.go' -not -path "./vendor/*"))
    gometalinter.v1 --config .linter.json
    bash <(curl -s https://codecov.io/bash)
fi
