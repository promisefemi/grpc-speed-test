proto-gen:
	protoc --go_out=cached/proto --go_opt=paths=source_relative --go-grpc_out=cached/proto --go-grpc_opt=paths=source_relative model.proto && protoc --go_out=client/proto --go_opt=paths=source_relative --go-grpc_out=client/proto --go-grpc_opt=paths=source_relative model.proto
