package core

import "github.com/everystreet/go-proj/v8/proj"

// ToEPSG4326 converts coordinate from given coordinate system definition to EPSG4326
// Definition can be obtained using following url https://epsg.io/<EPSG_CODE>.proj4
// For example https://epsg.io/31257.proj4
// Human-friendly page is https://epsg.io/31257
// Coordinates are rounded to 5-digits precision (~1 meter) (https://en.wikipedia.org/wiki/Decimal_degrees)
func ToEPSG4326(x, y float64, projDefinition string) (float64, float64, error) {
	coord := proj.XY{X: x, Y: y}

	err := proj.CRSToCRS(
		projDefinition,
		"EPSG:4326",
		func(pj proj.Projection) {
			proj.TransformForward(pj, &coord)
			// transform more coordinates
		},
	)

	if err != nil {
		return 0, 0, err
	}
	return TruncCoord(coord.X), TruncCoord(coord.Y), nil
}
