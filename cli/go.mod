module github.com/mberwanger/admiral/cli

go 1.24.1

replace github.com/mberwanger/admiral/client => ../client

replace github.com/mberwanger/admiral/server => ../server

require (
	github.com/bombsimon/logrusr/v2 v2.0.1
	github.com/go-logr/logr v1.4.2
	github.com/mberwanger/admiral/server v0.0.0-20250325173117-afa479381f7e
	github.com/muesli/mango-cobra v1.2.0
	github.com/muesli/roff v0.1.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/muesli/mango v0.1.0 // indirect
	github.com/muesli/mango-pflag v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
