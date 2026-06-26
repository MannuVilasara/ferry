.PHONY: deps build release clean install uninstall

build:
	@echo "=> Running checks and warnings..."
	@go vet ./...
	@echo "=> Building Ferry (Debug)..."
	@mkdir -p bin
	@go build -gcflags="all=-N -l" -o bin/ferry ./cmd/ferry
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

install: release
	@echo "=> Creating ferry system user and group..."
	@id -g ferry >/dev/null 2>&1 || groupadd --system ferry
	@id -u ferry >/dev/null 2>&1 || useradd --system -g ferry --no-create-home --shell /bin/false ferry
	@echo "=> Adding current user to ferry group..."
	@if [ -n "$${SUDO_USER}" ]; then usermod -aG ferry $${SUDO_USER}; else usermod -aG ferry $${USER}; fi
	@echo "=> Installing binary to /usr/local/bin..."
	@cp bin/ferry /usr/local/bin/ferry
	@chmod 755 /usr/local/bin/ferry
	@echo "=> Installing config to /etc/ferry..."
	@mkdir -p /etc/ferry
	@cp config.yaml /etc/ferry/config.yaml
	@chown -R ferry:ferry /etc/ferry
	@chmod 640 /etc/ferry/config.yaml
	@echo "=> Installing systemd service..."
	@cp ferry.service /etc/systemd/system/ferry.service
	@systemctl daemon-reload
	@systemctl enable --now ferry
	@echo "=> Install complete! (You may need to run 'newgrp ferry' or restart your terminal to use the CLI)"

uninstall:
	@echo "=> Stopping and disabling systemd service..."
	@systemctl stop ferry || true
	@systemctl disable ferry || true
	@rm -f /etc/systemd/system/ferry.service
	@systemctl daemon-reload
	@echo "=> Removing binary and config..."
	@rm -f /usr/local/bin/ferry
	@rm -rf /etc/ferry
	@echo "=> Removing system user and group..."
	@userdel ferry || true
	@groupdel ferry || true
	@echo "=> Uninstall complete!"
