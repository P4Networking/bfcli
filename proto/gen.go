package proto

//go:generate echo Generating pipe protobuf
//go:generate protoc pipe.proto --go_out=plugins=grpc:. -I.



