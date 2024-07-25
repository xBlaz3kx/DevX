proto:
	protoc --go_out=./proto \
	--go_opt=paths=source_relative \
  	--proto_path=./proto \
	--experimental_allow_proto3_optional \
    ./proto/*.proto

test:
	go test -v --short ./...

gen-mocks:
	 mockery .
