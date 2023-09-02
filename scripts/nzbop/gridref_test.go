package nzbop

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestConvertNZMS260(t *testing.T) {
	actual, err := convertNZMS260("T13: 6546 0652")
	expected := &core.Location{
		Latitude:  -37.50676,
		Longitude: 175.88701,
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
