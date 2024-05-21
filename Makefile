.DEFAULT_GOAL := build

TARGET = courier
VERSION = $$(make -s version)
VERSION_PATH = ./cmd/
BUILD_DIR = ./cmd/build/
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
FLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
GOBIN ?= $(shell go env GOPATH)/bin

OS_LIST = windows linux
ARCH_LIST = amd64 arm64

.PHONY: format
format:
	gofmt -l -s -w .

.PHONY: build
build:
	go build -ldflags=${FLAGS} -o ${TARGET} ./cmd/main.go && mv ${TARGET} ${BUILD_DIR}

.PHONY: clean
clean:
	@rm -rf ${BUILD_DIR}courier*
	@go clean

.PHONY: test
test:
	go test -v -race ./...

.PHONY: compress
compress:
	@if [ "$$(which upx)" != "" ]; then \
		upx -9 -k $(BUILD_DIR)$(TARGET); \
	else \
		echo "upx is not installed. Skipping compression."; \
	fi

.PHONY: lint
lint: $(GOBIN)/staticcheck
	@go vet ./...
	@$(GOBIN)/staticcheck -checks all ./...

.PHONY: version
version: $(GOBIN)/gobump
	 @$(GOBIN)/gobump show -r "$(VERSION_PATH)"

.PHONY: cross
cross:
	@for os in $(OS_LIST); do \
		for arch in $(ARCH_LIST); do \
			GOOS=$$os GOARCH=$$arch $(MAKE) build; \
			$(MAKE) compress; \
			rm ${BUILD_DIR}${TARGET}.~; \
			mkdir -p $(BUILD_DIR)$(TARGET)-$$os-$$arch; \
			mv $(BUILD_DIR)$(TARGET) $(BUILD_DIR)$(TARGET)-$$os-$$arch; \
			cp README* $(BUILD_DIR)$(TARGET)-$$os-$$arch; \
			cp LICENSE* $(BUILD_DIR)$(TARGET)-$$os-$$arch; \
			if [ "$$os" = "windows" ] || [ "$$os" = "darwin" ]; then \
				zip -r $(BUILD_DIR)$(TARGET)-$$os-$$arch.zip $(BUILD_DIR)$(TARGET)-$$os-$$arch >/dev/null 2>&1; \
			else \
				tar -czf $(BUILD_DIR)$(TARGET)-$$os-$$arch.tar.gz -C $(BUILD_DIR) $(TARGET)-$$os-$$arch >/dev/null 2>&1; \
			fi; \
			rm -r $(BUILD_DIR)$(TARGET)-$$os-$$arch; \
		done \
	done

$(GOBIN)/gobump:
	go install github.com/xoebus/gobump/cmd/gobump@latest

$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
