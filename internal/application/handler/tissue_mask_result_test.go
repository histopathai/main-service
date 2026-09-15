package handler

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testdata/tissue_mask.json was written by image-processing-service
// (himgproc --task tissue) for an NDPI slide.
func TestParseWorkerTissueMask(t *testing.T) {
	f, err := os.Open("testdata/tissue_mask.json")
	require.NoError(t, err)
	defer f.Close()

	data, err := parseWorkerTissueMask(f)
	require.NoError(t, err)

	details, ok := data.Validate()
	require.True(t, ok, "worker output must pass API validation: %v", details)
	assert.Equal(t, "tissue-v1", data.AlgorithmVersion)
	assert.Equal(t, "saturation-gray", data.Params.Method)
	assert.Equal(t, 2048, data.PreviewWidth)
	assert.Equal(t, 99840, data.Level0Width)
	assert.NotEmpty(t, data.Polygons)
	assert.GreaterOrEqual(t, len(data.Polygons[0].Exterior), 3)
}

func TestParseWorkerTissueMaskRejectsGarbage(t *testing.T) {
	_, err := parseWorkerTissueMask(strings.NewReader("{not json"))
	assert.Error(t, err)
}
