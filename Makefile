gen_proto:
	find proto -name "*.proto" -exec protoc --proto_path=proto --go_out=common/protocol/proto --go-grpc_out=.. --go_opt=paths=source_relative {} +