package main

import "math"

type Vec2 struct {
	X float64
	Y float64
}

func (v Vec2) Add(o Vec2) Vec2 {
	return Vec2{v.X + o.X, v.Y + o.Y}
}

func (v Vec2) Sub(o Vec2) Vec2 {
	return Vec2{v.X - o.X, v.Y - o.Y}
}

func (v Vec2) Scale(s float64) Vec2 {
	return Vec2{v.X * s, v.Y * s}
}

func (v Vec2) Length() float64 {
	return math.Hypot(v.X, v.Y)
}

func (v Vec2) Normalized() Vec2 {
	var length = v.Length()
	if length == 0 {
		return Vec2{}
	}

	return Vec2{v.X / length, v.Y / length}
}

func (v Vec2) DistanceTo(o Vec2) float64 {
	return o.Sub(v).Length()
}
