package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordinateToTimezone(t *testing.T) {
	assert := assert.New(t)
	defer CloseTimezoneDb()
	tz, err := CoordinateToTimezone(55.75222, 37.61556)
	if assert.NoError(err) {
		assert.Equal(tz, "Europe/Moscow")
	}
}
