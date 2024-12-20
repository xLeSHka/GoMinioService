PROTO_DIR := ../proto
OUT_DIR := ./pkg/api

MAIN_DIR := ./cmd/main

BINARY_NAME := fileservice

PROTOC := protoc
GO := go

PROTO_FILES := ${PROTO_DIR}/file/file.proto

all: generate build run

generate: ${PROTO_FILES}
	$(PROTOC)  \
    --go_out $(OUT_DIR) --go_opt paths=source_relative \
    --go-grpc_out $(OUT_DIR) --go-grpc_opt paths=source_relative \
    --proto_path=${PROTO_DIR} $(PROTO_FILES)

build: generate
	@echo "Build the binary..."
	${GO} build -o ${BINARY_NAME} ${MAIN_DIR}

run: build
	@echo "Running the server..."
	./${BINARY_NAME}

clean:
	@echo "Cleaning up..."
	rm -f ${BINARY_NAME}

.PHONY: all generate build run clean