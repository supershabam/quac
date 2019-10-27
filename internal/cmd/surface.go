package cmd

import (
	"context"

	"github.com/supershabam/quac/internal/quac"
	"gopkg.in/alecthomas/kingpin.v2"
)

func Surface(ctx context.Context, app *kingpin.Application) error {
	cmd := app.Command("surface", "allow network ports to be made accessible over quic")
	ports := cmd.Flag("port", "host:port").Strings()
	target := cmd.Flag("via", "address to dial").Default("quac.supershabam.com:443").String()
	cmd.Action(func(*kingpin.ParseContext) error {
		s := &quac.Surfacer{
			Target: *target,
			Ports:  *ports,
		}
		return s.Run(ctx)
	})
	return nil
}
