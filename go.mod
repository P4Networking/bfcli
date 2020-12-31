module github.com/P4Networking/pisc-cli

go 1.14

replace (
	github.com/P4Networking/pisc => ../pisc
	github.com/P4Networking/proto => ../proto
)

require (
	github.com/P4Networking/pisc v0.0.0-00010101000000-000000000000
	github.com/P4Networking/proto v1.0.0
	github.com/cheggaaa/pb/v3 v3.0.5
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	google.golang.org/grpc v1.33.1
)
