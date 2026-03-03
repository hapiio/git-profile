package main

import "github.com/hapiio/git-profile/cmd"

// ldflags to set version, commit, and date at build time
//
//	-X main.version=v1.2.3
//	-X main.commit=abc1234
//	-X main.date=2024-01-01T00:00:00Z
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
