module github.com/sonix-framework/sonix

go 1.26.1

require github.com/sonix-framework/core v0.0.0-00010101000000-000000000000

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/yaml v1.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.1 // indirect
	github.com/knadh/koanf/providers/env v1.1.0 // indirect
	github.com/knadh/koanf/providers/file v1.2.1 // indirect
	github.com/knadh/koanf/v2 v2.3.6 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)

// replace lokal untuk development; ganti dengan versi tag saat rilis (M6).
replace github.com/sonix-framework/core => ../core
