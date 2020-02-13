package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertCoordinates(t *testing.T) {
	assert := assert.New(t)
	epsg21781 := "+proj=somerc +lat_0=46.95240555555556 +lon_0=7.439583333333333 +k_0=1 +x_0=600000 +y_0=200000 +ellps=bessel +towgs84=674.4,15.1,405.3,0,0,0,0 +units=m +no_defs"
	// Check here https://epsg.io/transform#s_srs=21781&t_srs=4326&x=575500.0000000&y=197790.0000000
	x, y, err := ToEPSG4326(575500, 197790, epsg21781)
	if assert.NoError(err) {
		assert.Equal(7.11691, x)
		assert.Equal(46.93074, y)
	}
	_, _, e := ToEPSG4326(575500, 197790, "junk")
	assert.Error(e)
}
