PLUGIN_EXE:=plugin/plumgrid

.DEFAULT: all
.PHONY: all

PLUGIN_VERSION=git-$(shell git rev-parse --short=12 HEAD)

$(PLUGIN_EXE): plugin/main.go plugin/driver/*.go
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\" -X main.version=$(PLUGIN_VERSION)" -tags netgo -o $@ ./$(@D)
	@strings $@ | grep cgo_stub\\\.go >/dev/null || { \
		rm $@; \
		echo "\nGo standard library is built without the 'netgo' build tag."; \
		echo "Please run"; \
		echo "    sudo go clean -i net"; \
		echo "    sudo go install -tags netgo std"; \
		false; \
	}
