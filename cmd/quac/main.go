package main

import (
	"context"
	"os"

	"github.com/supershabam/quac/internal/cmd"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// Version is injected upon build
	Version string

	app         = kingpin.New("quac", "quic underlay authenticated connections")
	subcommands = []func(context.Context, *kingpin.Application) error{
		cmd.Dial,
		cmd.Serve,
		cmd.Surface,
	}
)

func main() {
	err := run(context.Background())
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("unhandled error")
	}
}

func run(ctx context.Context) error {
	// TODO allow log level and style to be specified
	l, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(l)
	app.Version(Version)
	for _, cmd := range subcommands {
		err := cmd(ctx, app)
		if err != nil {
			return err
		}
	}
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	return nil
}
