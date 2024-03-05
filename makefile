GOCMD=go
GOTEST=$(GOCMD) test

.PHONY: test build dep

# Project output directory.
PWD := $(shell pwd)

OUTPUT_DIR := $(PWD)/bin

TARGETS := $(shell basename -s .go ./cli/*.go)

echo:
	@echo $(TARGETS)

build: dep
	@for target in $(TARGETS); do \
    	cd $(PWD)/cli; $(GOCMD) build -o ../bin/$$target ./$$target.go ; \
    done

web: dep
	@mkdir -p ./bin
	@cd $(PWD)/web; $(GOCMD) build -o ../bin/www web.go

dep:
	@cd ./internal; $(GOCMD) mod tidy
	@cd ./task; $(GOCMD) mod tidy
	@cd ./cli; $(GOCMD) mod tidy
	@cd ./task; $(GOCMD) mod tidy

clean:
	@rm -f $(OUTPUT_DIR)/*
	@cd ./internal; $(GOCMD) clean -modcache
	@cd ./task; $(GOCMD) clean -modcache
	@cd ./cli; $(GOCMD) clean -modcache
	@cd ./test; $(GOCMD) clean -modcache
	@$(GOCMD) clean

test:
	@cd ./test; $(GOTEST)