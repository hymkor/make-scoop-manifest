NAME=$(lastword $(subst /, ,$(abspath .)))
VERSION=$(shell git.exe describe --tags 2>nul || echo noversion)
GOOPT=-ldflags "-s -w -X main.version=$(VERSION)"

ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=SET
else
    SET=export
endif

all:
	go fmt
	$(SET) "CGO_ENABLED=0" && go build $(GOOPT)

_package:
	go fmt
	$(SET) "CGO_ENABLED=0" && go build $(GOOPT)
	zip $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXT)

package:
	$(SET) "GOOS=linux"   && $(SET) "GOARCH=386"   && $(MAKE) _package EXT=
	$(SET) "GOOS=linux"   && $(SET) "GOARCH=amd64" && $(MAKE) _package EXT=
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _package EXT=.exe
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _package EXT=.exe

manifest:
	go run ./mkmanifest.go *-windows-*.zip > $(NAME).json
