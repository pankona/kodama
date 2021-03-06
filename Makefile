
MAKEFLAGS += --no-builtin-rules

PROTO_DIR = proto
PB_GO_DIR = service

PROTOS = $(shell find $(PROTO_DIR) -printf "%f\n" | grep proto$$)
PB_GOS = $(PROTOS:%.proto=$(PB_GO_DIR)/%.pb.go)

BUILD_CMD ?= go build

DIRECTORIES = $(dir $(wildcard $(CURDIR)/cmd/*/.))

all: build

build: $(DIRECTORIES)

$(DIRECTORIES): genproto
	cd $@; go build

genproto: $(PB_GOS)

$(PB_GO_DIR)/%.pb.go: $(PROTO_DIR)/%.proto
	mkdir -p $(dir $@)
	protoc -I $(PROTO_DIR) --go_out=plugins=grpc:$(dir $@) ./$<

clean:
	rm -rf $(PB_GO_DIR)
