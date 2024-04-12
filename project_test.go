package sgunk

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProject(t *testing.T) {
	projectFS := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1")
	config, err := LoadConfigFile(projectFS)
	require.NoError(t, err)

	project := New(
		WithWorkDir("testdata/project1"),
		WithConfig(config),
	)

	err = project.Generate()
	assert.NoError(t, err)
}
