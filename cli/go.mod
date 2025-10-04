module github.com/cline/cli

go 1.23.0

require (
	github.com/cline/grpc-go v0.0.0
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/spf13/cobra v1.8.0
	google.golang.org/grpc v1.75.0
)

replace github.com/cline/grpc-go => ../src/generated/grpc-go

require (
	github.com/AlecAivazis/survey/v2 v2.3.7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
