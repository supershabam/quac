package quac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	Host string

	sessions map[string]quic.Session
}

func (s *Server) Run(ctx context.Context) error {
	s.sessions = map[string]quic.Session{}
	eg, ctx := errgroup.WithContext(ctx)
	m := &autocert.Manager{
		Cache:      autocert.DirCache("secret-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.Host),
		Email:      "ian@supershabam.com",
	}
	eg.Go(func() error {
		hs := &http.Server{
			Addr:      ":https",
			TLSConfig: m.TLSConfig(),
		}
		eg.Go(func() error {
			<-ctx.Done()
			tctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			return hs.Shutdown(tctx)
		})
		err := hs.ListenAndServeTLS("", "")
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("while listening to https server: %w", err)
		}
		return err
	})
	eg.Go(func() error {
		cfg := m.TLSConfig()
		cfg.NextProtos = []string{"quac-v1"}
		l, err := quic.ListenAddr(":443", cfg, nil)
		if err != nil {
			return fmt.Errorf("while listening to quic server: %w", err)
		}
		for {
			sess, err := l.Accept(ctx)
			if err != nil {
				return fmt.Errorf("while accepting: %w", err)
			}
			go s.handle(ctx, sess)
		}
	})
	return eg.Wait()
}

// a session may be a surfacer or a dialer
func (s *Server) handle(ctx context.Context, sess quic.Session) {
	for {
		stream, err := sess.AcceptStream(ctx)
		if err != nil {
			return
		}
		go s.stream(ctx, sess, stream)
	}
}

func (s *Server) stream(ctx context.Context, sess quic.Session, stream quic.Stream) {
	var dialerOrSurfacer struct {
		Ports []string `json:"ports"`
		Dial  string   `json:"dial"`
	}
	r, err := reader(stream, &dialerOrSurfacer)
	if err != nil {
		return
	}
	if dialerOrSurfacer.Dial != "" {
		zap.L().Info("handling dialer")
		// handler dialer
		for port, sess := range s.sessions {
			if dialerOrSurfacer.Dial != port {
				continue
			}
			zap.L().With(
				zap.String("port", port),
			).Info("joining dialer to surfacer")
			dest, err := sess.OpenStream()
			if err != nil {
				stream.Close()
				return
			}
			b, err := json.Marshal(dialerOrSurfacer)
			if err != nil {
				return
			}
			err = writer(dest, b)
			if err != nil {
				return
			}
			eg, ctx := errgroup.WithContext(ctx)
			eg.Go(func() error {
				_, err := io.Copy(dest, r)
				return err
			})
			eg.Go(func() error {
				_, err := io.Copy(stream, dest)
				return err
			})
			eg.Go(func() error {
				<-ctx.Done()
				return dest.Close()
			})
			eg.Go(func() error {
				<-ctx.Done()
				return stream.Close()
			})
		}
		stream.Close()
	} else if len(dialerOrSurfacer.Ports) > 0 {
		// handle surfacer
		zap.L().Info("handling surfacer")
		for _, port := range dialerOrSurfacer.Ports {
			s.sessions[port] = sess // this is wrong, but should get data flowing
		}
	} else {
		zap.L().Info("closing stream because unknown first message")
		stream.Close()
	}
}
