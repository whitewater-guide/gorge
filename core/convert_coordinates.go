package core

import "github.com/whitewater-guide/gorge/proj"

const epsg4326Def = "+proj=longlat +datum=WGS84 +no_defs"

// ToEPSG4326 converts coordinate from given coordinate system definition to EPSG4326
// Definition can be obtained using following url https://epsg.io/<EPSG_CODE>.proj4
// For example https://epsg.io/31257.proj4
// Human-friendly page is https://epsg.io/31257
// Coordinates are rounded to 5-digits precision (~1 meter) (https://en.wikipedia.org/wiki/Decimal_degrees)
func ToEPSG4326(x, y float64, projDefinition string) (float64, float64, error) {
	ctx := proj.NewContext()
	defer ctx.Close()

	fromProj, err := ctx.Create(projDefinition)
	if err != nil {
		return 0, 0, err
	}
	defer fromProj.Close()

	epsg4326, err := ctx.Create(epsg4326Def)
	if err != nil {
		return 0, 0, err
	}
	defer epsg4326.Close()

	xFrom, yFrom, _, _, err := fromProj.Trans(proj.Inv, x, y, 0, 0)
	if err != nil {
		return 0, 0, err
	}
	xTo, yTo, _, _, err := epsg4326.Trans(proj.Fwd, xFrom, yFrom, 0, 0)
	return TruncCoord(proj.RadToDeg(xTo)), TruncCoord(proj.RadToDeg(yTo)), err
}
