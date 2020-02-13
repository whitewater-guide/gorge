package quebec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuebec_ConvertDMS(t *testing.T) {
	assert := assert.New(t)
	actual, err := convertDMS("59°51'48\"")
	if assert.NoError(err) {
		assert.Equal(59.86333, actual)
	}
}
