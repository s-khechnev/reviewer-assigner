package integration_tests

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed fixtures
var Fixtures embed.FS

type FixtureLoader struct {
	t           *testing.T
	currentPath fs.FS
}

func NewFixtureLoader(t *testing.T, fixturePath fs.FS) *FixtureLoader {
	return &FixtureLoader{
		t:           t,
		currentPath: fixturePath,
	}
}

func (l *FixtureLoader) LoadString(path string) string {
	file, err := l.currentPath.Open(path)
	require.NoError(l.t, err)

	defer file.Close()

	data, err := io.ReadAll(file)
	require.NoError(l.t, err)

	return string(data)
}

func (l *FixtureLoader) LoadTemplate(s string, data any) string {
	const defaultName = "default"

	temp, err := template.New(defaultName).Parse(s)
	require.NoError(l.t, err)

	buf := bytes.Buffer{}

	err = temp.Execute(&buf, data)
	require.NoError(l.t, err)

	return buf.String()
}
