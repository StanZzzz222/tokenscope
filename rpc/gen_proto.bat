@echo off
echo "Generate protobuf file..."
if not exist pb mkdir pb
protoc --go_out=./ --go-grpc_out=./ proto/*.proto --proto_path=proto
pause
