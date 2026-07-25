package main

import "math"

type OBB struct {
	Center     Vector2
	HalfWidth  float64
	HalfHeight float64
	Rot        float64
}

func (b OBB) Axes() (Vector2, Vector2) {
	var c = math.Cos(b.Rot)
	var s = math.Sin(b.Rot)
	return Vector2{X: c, Y: s}, Vector2{X: -s, Y: c}
}

func (b OBB) Corners() [4]Vector2 {
	var fwd, side = b.Axes()
	var x = fwd.Scale(b.HalfWidth)
	var y = side.Scale(b.HalfHeight)
	return [4]Vector2{
		b.Center.Add(x).Add(y),
		b.Center.Sub(x).Add(y),
		b.Center.Sub(x).Sub(y),
		b.Center.Add(x).Sub(y),
	}
}

func (b OBB) Overlap(other OBB) (Vector2, bool) {
	var bCorners = b.Corners()
	var oCorners = other.Corners()

	var bFwd, bSide = b.Axes()
	var oFwd, oSide = other.Axes()
	var axes = [4]Vector2{bFwd, bSide, oFwd, oSide}

	var minOverlap = math.Inf(1)
	var separationAxis Vector2

	for _, axis := range axes {
		var bMin, bMax = projectCorners(bCorners, axis)
		var oMin, oMax = projectCorners(oCorners, axis)

		var overlap = math.Min(bMax, oMax) - math.Max(bMin, oMin)
		if overlap <= 0 {
			return Vector2{}, false
		}

		if overlap < minOverlap {
			minOverlap = overlap
			separationAxis = axis
		}
	}

	// Orient the separation so it points from b toward other.
	if other.Center.Sub(b.Center).Dot(separationAxis) < 0 {
		separationAxis = separationAxis.Scale(-1)
	}

	return separationAxis.Scale(minOverlap), true
}

// projectCorners returns the [min, max] shadow of the corners on a unit axis.
func projectCorners(corners [4]Vector2, axis Vector2) (float64, float64) {
	var minP = corners[0].Dot(axis)
	var maxP = minP
	for i := 1; i < 4; i++ {
		var p = corners[i].Dot(axis)
		minP = math.Min(minP, p)
		maxP = math.Max(maxP, p)
	}
	return minP, maxP
}
