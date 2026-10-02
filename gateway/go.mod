module github.com/Vlad777-bit/personal-finance-analytics/gateway

go 1.26

require (
	github.com/Vlad777-bit/personal-finance-analytics/shared v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.12.1
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)

replace github.com/Vlad777-bit/personal-finance-analytics/shared => ../shared
