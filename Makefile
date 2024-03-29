BIN := gyrolock
VERSION := 0.1

PREFIX ?= /usr
LIB_DIR = $(DESTDIR)$(PREFIX)/lib
BIN_DIR = $(DESTDIR)$(PREFIX)/bin
SHARE_DIR = $(DESTDIR)$(PREFIX)/share

export CGO_CPPFLAGS := ${CPPFLAGS}
export CGO_CFLAGS := ${CFLAGS}
export CGO_CXXFLAGS := ${CXXFLAGS}
export CGO_LDFLAGS := ${LDFLAGS}
export GOFLAGS := -buildmode=pie -trimpath -mod=readonly -modcacherw

.PHONY: build
build: main.go
	go build -o $(BIN) main.go

.PHONY: release
release: build
	strip $(BIN) 2>/dev/null || true
	upx -9 $(BIN) 2>/dev/null || true

.PHONY: install
install:
	install -Dm755 -t "$(BIN_DIR)/" $(BIN)
	install -Dm644 -t "$(SHARE_DIR)/licenses/$(BIN)/" LICENSE.md
	install -Dm644 -t "$(LIB_DIR)/systemd/system/" $(BIN).service
	install -Dm644 -t "$(LIB_DIR)/udev/rules.d/" 80-$(BIN).rules
	install -Dm644 -t "$(LIB_DIR)/systemd/user/" swaylock.service

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

.PHONY: clean
clean:
	rm -f "$(BIN)"
	rm -rf dist
	rm -rf vendor

.PHONY: dist
dist: clean vendor build
	$(eval TMP := $(shell mktemp -d))
	mkdir "$(TMP)/$(BIN)-$(VERSION)"
	cp -r * "$(TMP)/$(BIN)-$(VERSION)"
	(cd "$(TMP)" && tar -cvzf "$(BIN)-$(VERSION)-src.tar.gz" "$(BIN)-$(VERSION)")

	mkdir "$(TMP)/$(BIN)-$(VERSION)-linux64"
	cp LICENSE.md $(BIN) $(BIN).service 80-$(BIN).rules swaylock.service "$(TMP)/$(BIN)-$(VERSION)-linux64"
	(cd "$(TMP)" && tar -cvzf "$(BIN)-$(VERSION)-linux64.tar.gz" "$(BIN)-$(VERSION)-linux64")

	mkdir -p dist
	mv "$(TMP)/$(BIN)-$(VERSION)"-*.tar.gz dist
	git archive -o "dist/$(BIN)-$(VERSION).tar.gz" --format tar.gz --prefix "$(BIN)-$(VERSION)/" "v$(VERSION)"

	for file in dist/*; do \
	    gpg --detach-sign --armor "$$file"; \
	done

	rm -rf "$(TMP)"
	rm -f "dist/$(BIN)-$(VERSION).tar.gz"
