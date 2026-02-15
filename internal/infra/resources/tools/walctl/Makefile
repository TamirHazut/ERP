# walctl CLI tool Makefile

.PHONY: build clean install

BIN_DIR := ../bin
BINARY := walctl

build:
	@echo "Building walctl..."
	@go build -o $(BIN_DIR)/$(BINARY) .
	@echo "✓ walctl built successfully at $(BIN_DIR)/$(BINARY)"

clean:
	@echo "Cleaning walctl binary..."
	@rm -f $(BIN_DIR)/$(BINARY)
	@echo "✓ Clean complete"

install: build
	@echo "Installing walctl to /usr/local/bin..."
	@cp $(BIN_DIR)/$(BINARY) /usr/local/bin/
	@echo "✓ walctl installed successfully"

help:
	@echo "walctl Makefile targets:"
	@echo "  make build    - Build the walctl binary"
	@echo "  make clean    - Remove the binary"
	@echo "  make install  - Build and install to /usr/local/bin"
	@echo "  make help     - Show this help"
