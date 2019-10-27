package quac

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReader(t *testing.T) {
	const jsonpart = `{"something":"else"}` + "\n"
	buf := make([]byte, binary.MaxVarintLen32)
	n := binary.PutVarint(buf, int64(len(jsonpart)))
	buf = buf[:n]
	input := io.MultiReader(bytes.NewReader(buf), strings.NewReader(jsonpart+`and then some more`))
	var m map[string]string
	output, err := reader(input, &m)
	require.Nil(t, err)
	b, err := ioutil.ReadAll(output)
	require.Nil(t, err)
	if !bytes.Equal(b, []byte("and then some more")) {
		t.Errorf("unexpected remainder of reader: %s", b)
	}
	if !reflect.DeepEqual(m, map[string]string{
		"something": "else",
	}) {
		t.Errorf("unexpected json message: %v", m)
	}
}
