module github.com/yhonda-ohishi/etc_meisai

go 1.24.0

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.2
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.11.1
	golang.org/x/text v0.29.0
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/db_service => C:/go/db_service
