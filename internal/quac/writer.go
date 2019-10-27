package quac

import (
	"bytes"
	"encoding/binary"
	"io"
)

func writer(w io.Writer, buf []byte) error {
	lbuf := make([]byte, binary.MaxVarintLen32)
	n := binary.PutVarint(lbuf, int64(len(buf)))
	lbuf = lbuf[:n]
	_, err := io.Copy(w, io.MultiReader(bytes.NewReader(lbuf), bytes.NewReader(buf)))
	return err
}
