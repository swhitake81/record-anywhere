PREFIX ?= /usr/local
BINARY = record-anywhere

build:
	go build -o $(BINARY) .

install: build
	install -d $(PREFIX)/bin
	install -m 755 $(BINARY) $(PREFIX)/bin/

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

deps:
	brew install portaudio ffmpeg
	brew install --cask blackhole-2ch

setup: deps install

clean:
	rm -f $(BINARY)

.PHONY: build install uninstall deps setup clean
