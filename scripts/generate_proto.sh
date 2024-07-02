protoc --go_out=plugin/proto/src/golang/gcp --go_opt=paths=source_relative \
    --go-grpc_out=plugin/proto/src/golang/gcp --go-grpc_opt=paths=source_relative \
    plugin/proto/*.proto
mv plugin/proto/src/golang/gcp/plugin/proto/* plugin/proto/src/golang/gcp
rm -rf plugin/proto/src/golang/gcp/plugin