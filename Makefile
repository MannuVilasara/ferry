.PHONY: deps build release clean

build:
	@echo "=> Building Ferry (Debug)..."
	@mkdir -p bin
	@go build -o bin/ferry ./cmd/ferry
	@echo "=> Build complete! Run ./bin/ferry to start it."

deps:
	@echo "=> Downloading dependencies..."
	@go mod tidy

release:
	@echo "=> Building Ferry (Release)..."
	@mkdir -p bin
	@go build -ldflags="-s -w" -o bin/ferry ./cmd/ferry
	@echo "=> Release build complete!"

clean:
	@echo "=> Cleaning up..."
	@rm -rf bin/
	@rm -f /tmp/ferry.pid
	@rm -f /tmp/ferry.sock
