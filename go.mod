module github.com/akabos/actions-runner

go 1.14

replace github.com/google/go-github/v31 v31.0.0 => github.com/google/go-github/v31 v31.0.0-20200424141832-6a666c5177f5

require (
	github.com/bradleyfalzon/ghinstallation v1.1.1
	github.com/google/go-github/v31 v31.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.8.0
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
)
