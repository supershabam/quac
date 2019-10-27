package quac

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/davecgh/go-spew/spew"
	quic "github.com/lucas-clemente/quic-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Dialer struct {
	Target string
	Addr   string
}

func (d *Dialer) Run(ctx context.Context) error {
	sess, err := quic.DialAddrContext(ctx, d.Target, &tls.Config{
		NextProtos: []string{"quac-v1"},
	}, nil)
	if err != nil {
		return fmt.Errorf("while dialing: %w", err)
	}
	stream, err := d.registerDial(ctx, sess)
	if err != nil {
		return err
	}
	return d.read(ctx, stream)
}

func (d *Dialer) registerDial(ctx context.Context, sess quic.Session) (quic.Stream, error) {
	stream, err := sess.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	req := struct {
		Dial string `json:"dial"`
	}{
		Dial: d.Addr,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = writer(stream, b)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (d *Dialer) read(ctx context.Context, stream quic.Stream) error {
	eg, ctx := errgroup.WithContext(ctx)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	spew.Dump(l.Addr())
	eg.Go(func() error {
		<-ctx.Done()
		return l.Close()
	})
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	zap.L().Info("accepted connection")
	eg.Go(func() error {
		<-ctx.Done()
		return conn.Close()
	})
	eg.Go(func() error {
		<-ctx.Done()
		return stream.Close()
	})
	eg.Go(func() error {
		_, err := io.Copy(conn, stream)
		return err
	})
	eg.Go(func() error {
		_, err := io.Copy(stream, conn)
		return err
	})
	return eg.Wait()
}
