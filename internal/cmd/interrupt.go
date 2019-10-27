package cmd

import (
	"context"
	"os"
	"os/signal"
)

func interrupt(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
