package cmd

import (
	"context"

	"github.com/supershabam/quac/internal/quac"
	"gopkg.in/alecthomas/kingpin.v2"
)

func Dial(ctx context.Context, app *kingpin.Application) error {
	cmd := app.Command("dial", "dial a surfaced port")
	via := cmd.Flag("via", "quac server by which to initiate connectivity").Default("quac.supershabam.com:443").String()
	addr := cmd.Flag("addr", "address on remote server to dial").Required().String()
	cmd.Action(func(*kingpin.ParseContext) error {
		d := &quac.Dialer{
			Target: *via,
			Addr:   *addr,
		}
		return d.Run(ctx)
	})
	return nil
}
