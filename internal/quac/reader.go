package quac

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"

	"go.uber.org/zap"
)

func reader(r io.Reader, v interface{}) (io.Reader, error) {
	br := bufio.NewReader(r)
	l, err := binary.ReadVarint(br)
	if err != nil {
		return nil, err
	}
	zap.L().With(
		zap.Int64("length", l),
	).Info("reading json message")
	buf := make([]byte, int(l))
	_, err = io.ReadFull(br, buf)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, v)
	if err != nil {
		return nil, err
	}
	zap.L().With(
		zap.ByteString("buf", buf),
	).Info("read message")
	nr := br.Buffered()
	read, err := br.Peek(nr)
	if err != nil {
		return nil, err
	}
	return io.MultiReader(bytes.NewReader(read), r), nil
}
