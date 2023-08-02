.PHONY: coveralls cover install-gotestsum test-report

coveralls:
	go test --timeout 120s -coverprofile=profile.cov -covermode=atomic -coverpkg=github.com/bjartek/overflow -v ./...

cover: coveralls
	go tool cover -html=profile.cov

install-gotestsum:
	go install gotest.tools/gotestsum@latest

test-report: install-gotestsum
	gotestsum -f testname --no-color --hide-summary failed --junitfile test-result.xml
