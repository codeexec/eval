package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

var (
	s = `// :run go run main.go -echo echo-arg additional arg
// :collection Essential Go
package main

import (
	"flag"
	"fmt"
	"os"
)
`
)

func TestParseMetaFromText(t *testing.T) {
	m := parseMetaFromText(s)
	assert.True(t, m.DidParse)
	assert.Equal(t, m.RunCmd, "go run main.go -echo echo-arg additional arg")
	assert.Equal(t, m.Collection, "Essential Go")
}
