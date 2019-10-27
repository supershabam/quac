package quac

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"

	quic "github.com/lucas-clemente/quic-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Surfacer struct {
	Target string
	Ports  []string
}

func (s *Surfacer) Run(ctx context.Context) error {
	sess, err := quic.DialAddrContext(ctx, s.Target, &tls.Config{
		NextProtos: []string{"quac-v1"},
	}, nil)
	if err != nil {
		return fmt.Errorf("while dialing: %w", err)
	}
	err = s.registerPorts(ctx, sess)
	if err != nil {
		return err
	}
	return s.read(ctx, sess)
}

type registerPortsMessage struct {
	Ports []string `json:"ports"`
}

func (s *Surfacer) registerPorts(ctx context.Context, sess quic.Session) error {
	stream, err := sess.OpenStreamSync(ctx)
	if err != nil {
		return err
	}
	b, err := json.Marshal(registerPortsMessage{
		Ports: s.Ports,
	})
	if err != nil {
		return err
	}
	err = writer(stream, b)
	if err != nil {
		return err
	}
	return nil
}

func (s *Surfacer) read(ctx context.Context, sess quic.Session) error {
	for {
		stream, err := sess.AcceptStream(ctx)
		if err != nil {
			return err
		}
		go s.stream(ctx, stream)
	}
}

type dialRequest struct {
	Port string `json:"port"`
}

func (s *Surfacer) stream(ctx context.Context, stream quic.Stream) {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		<-ctx.Done()
		return stream.Close()
	})
	eg.Go(func() error {
		var dialReq struct {
			Dial string `json:"dial"`
		}
		r, err := reader(stream, &dialReq)
		if err != nil {
			return err
		}
		zap.L().With(
			zap.String("addr", dialReq.Dial),
		).Info("dialing")
		conn, err := net.Dial("tcp", dialReq.Dial)
		if err != nil {
			return err
		}
		eg.Go(func() error {
			_, err := io.Copy(conn, r)
			return err
		})
		_, err = io.Copy(stream, conn)
		return err
	})
	err := eg.Wait()
	if err != nil {
		zap.L().With(zap.Error(err)).Info("stream encountered an error")
	}

}
