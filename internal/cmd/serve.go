package cmd

import (
	"context"

	"github.com/supershabam/quac/internal/quac"
	"gopkg.in/alecthomas/kingpin.v2"
)

func Serve(ctx context.Context, app *kingpin.Application) error {
	cmd := app.Command("serve", "serve a quac")
	host := cmd.Flag("host", "hostname to acme whitelist").Default("quac.supershabam.com").String()
	cmd.Action(func(*kingpin.ParseContext) error {
		ctx = interrupt(ctx)
		s := &quac.Server{
			Host: *host,
		}
		return s.Run(ctx)
	})
	return nil
}
