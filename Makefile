.PHONY: build install clean snapshot

build:
	go build -o grove .

install: build
	mv grove ~/bin/grove

clean:
	rm -f grove
	rm -rf dist/

snapshot:
	goreleaser release --snapshot --clean
