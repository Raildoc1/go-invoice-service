module go-invoice-service/api-service

go 1.24.0

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/jwtauth/v5 v5.3.3
	github.com/shopspring/decimal v1.4.0
	go-invoice-service/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx/v2 v2.1.3 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

replace go-invoice-service/common => ../common
