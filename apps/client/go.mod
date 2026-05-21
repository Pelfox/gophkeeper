module github.com/Pelfox/gophkeeper/apps/client

go 1.26.3

require (
	github.com/aquasecurity/table v1.11.0
	github.com/nyaosorg/go-box/v3 v3.1.1
	github.com/oapi-codegen/runtime v1.4.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/crypto v0.51.0
	golang.org/x/term v0.43.0
)

require (
	github.com/clipperhouse/uax29/v2 v2.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/mattn/go-tty v0.0.7 // indirect
	github.com/nyaosorg/go-ttyadapter v0.3.0 // indirect
)

require (
	github.com/Pelfox/gophkeeper/shared/protocol v0.0.0-00010101000000-000000000000
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.6.0
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/sys v0.44.0 // indirect
)

replace github.com/Pelfox/gophkeeper/shared/protocol => ../../shared/protocol
