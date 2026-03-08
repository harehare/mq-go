.PHONY: build-rust test clean

MQ_REPO_URL := https://github.com/harehare/mq
MQ_REPO_DIR := .mq
MQ_FFI_DIR  := $(MQ_REPO_DIR)/crates/mq-ffi
MQ_LIB_DIR  := $(MQ_REPO_DIR)/target/release

setup:
	@if [ -d "$(MQ_REPO_DIR)/.git" ]; then \
		git -C $(MQ_REPO_DIR) pull --ff-only -q; \
	else \
		git clone --depth=1 $(MQ_REPO_URL) $(MQ_REPO_DIR); \
	fi

build-rust:
	cd $(MQ_FFI_DIR) && cargo build --release

test: build-rust
	CGO_LDFLAGS="-L$(MQ_LIB_DIR)" \
	LD_LIBRARY_PATH=$(MQ_LIB_DIR) \
	DYLD_LIBRARY_PATH=$(MQ_LIB_DIR) \
	go test -v ./...

clean:
	go clean -testcache
	cargo clean

