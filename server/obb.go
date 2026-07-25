package main

import "math"

type OBB struct {
	Center  Vector2
	Extents Vector2
	Rot     float64
}

func (b OBB) Axes() (Vector2, Vector2) {
	var cos = math.Cos(b.Rot)
	var sin = math.Sin(b.Rot)
	return Vector2{X: cos, Y: sin}, Vector2{X: -sin, Y: cos}
}

func (b OBB) WorldToLocal(point Vector2) Vector2 {
	var forward, right = b.Axes()
	var offset = point.Sub(b.Center)
	return Vector2{X: offset.Dot(forward), Y: offset.Dot(right)}
}

func (b OBB) ContainsPoint(point Vector2, margin float64) bool {
	var local = b.WorldToLocal(point)
	return math.Abs(local.X) <= b.Extents.X+margin && math.Abs(local.Y) <= b.Extents.Y+margin
}

func (b OBB) IntersectsLineSegment(from Vector2, to Vector2, margin float64) bool {
	var localFrom = b.WorldToLocal(from)
	var localDir = b.WorldToLocal(to).Sub(localFrom)

	var tMin = 0.0
	var tMax = 1.0
	var overlaps bool

	if tMin, tMax, overlaps = clipSlab(localFrom.X, localDir.X, b.Extents.X+margin, tMin, tMax); !overlaps {
		return false
	}

	if _, _, overlaps = clipSlab(localFrom.Y, localDir.Y, b.Extents.Y+margin, tMin, tMax); !overlaps {
		return false
	}

	return true
}

// clipSlab narrows [tMin, tMax] to the part of the segment inside one axis's slab
// [-ext, ext]; ok is false once the surviving range is empty.
func clipSlab(origin float64, direction float64, extent float64, tMin float64, tMax float64) (float64, float64, bool) {
	const epsilon = 1e-9
	if math.Abs(direction) < epsilon {
		return tMin, tMax, origin >= -extent && origin <= extent
	}

	var tNear = (-extent - origin) / direction
	var tFar = (extent - origin) / direction
	if tNear > tFar {
		tNear, tFar = tFar, tNear
	}

	tMin = math.Max(tMin, tNear)
	tMax = math.Min(tMax, tFar)
	return tMin, tMax, tMin <= tMax
}

func (b OBB) Corners() [4]Vector2 {
	var forward, right = b.Axes()
	var forwardExtent = forward.Scale(b.Extents.X)
	var rightExtent = right.Scale(b.Extents.Y)
	return [4]Vector2{
		b.Center.Add(forwardExtent).Add(rightExtent),
		b.Center.Sub(forwardExtent).Add(rightExtent),
		b.Center.Sub(forwardExtent).Sub(rightExtent),
		b.Center.Add(forwardExtent).Sub(rightExtent),
	}
}

func (b OBB) Overlap(other OBB) (Vector2, bool) {
	var bCorners = b.Corners()
	var oCorners = other.Corners()

	var bForward, bRight = b.Axes()
	var oForward, oRight = other.Axes()
	var axes = [4]Vector2{bForward, bRight, oForward, oRight}

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
	var minProjection = corners[0].Dot(axis)
	var maxProjection = minProjection
	for i := 1; i < 4; i++ {
		var projection = corners[i].Dot(axis)
		minProjection = math.Min(minProjection, projection)
		maxProjection = math.Max(maxProjection, projection)
	}
	return minProjection, maxProjection
}
