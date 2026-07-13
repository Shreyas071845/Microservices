module tracking_service

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.12.3
	google.golang.org/grpc v1.65.0
	pb v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240711142825-46eb208f015d // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace pb => ../../pkg/pb
