# Deprecated: use buf generate instead of proto
proto:
	protoc --go_out=./proto \
	--go_opt=paths=source_relative \
  	--proto_path=./proto \
	--experimental_allow_proto3_optional \
    ./proto/*.proto

# Run the tests
test:
	go test -v --short ./...

# Autogenerate all files (mocks, Protobuf, etc.)
gen: buf-gen gen-mocks

# Generate mocks for testing using mockery
gen-mocks:
	 mockery .

# Generate the Protobuf files using Buf
buf-gen:
	buf generate

# Lint the proto files using Buf
buf-lint:
	protoc -I . --include_source_info $(find . -name '*.proto') -o /dev/stdout | buf lint -
