package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowVersion(t *testing.T) {
	// redirect stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showVersion()

	// restore stdout
	w.Close()
	os.Stdout = origStdout

	// read captured output
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	assert.NoError(t, err)
	out := buf.String()

	assert.Contains(t, string(out), "version: devel")
	assert.Contains(t, string(out), "git commit: 0000000000000000000000000000000000000000")
	assert.Contains(t, string(out), "date: ")
}
