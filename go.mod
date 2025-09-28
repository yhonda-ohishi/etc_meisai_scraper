module github.com/yhonda-ohishi/etc_meisai_scraper

go 1.24.0

replace github.com/db_service => C:/go/db_service

require (
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
)
