# cmd ディレクトリ配下の go package をすべて bin ディレクトリにビルドする
CMDS := $(shell find cmd -mindepth 1 -maxdepth 1 -type d)
TARGETS := $(CMDS:cmd/%=bin/%)

.PHONY: all
all: $(TARGETS)

bin/%:
	go build -o $@ ./cmd/$*
