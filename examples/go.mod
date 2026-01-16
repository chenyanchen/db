module github.com/chenyanchen/kv/examples

go 1.24

require (
	github.com/chenyanchen/kv v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.20.5
)

replace github.com/chenyanchen/kv => ../

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	golang.org/x/sys v0.27.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
