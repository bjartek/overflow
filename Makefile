.PHONY: coveralls cover install-gotestsum test-report

coveralls:
	go test -coverprofile=profile.cov -covermode=atomic -coverpkg=github.com/bjartek/overflow/overflow -v ./...

cover: test
	go tool cover -html=profile.cov

install-gotestsum:
	go get gotest.tools/gotestsum

test-report: install-gotestsum
	gotestsum -f testname --no-color --hide-summary failed --junitfile test-result.xml
