package main

import "math"

type Vector2 struct {
	X float64
	Y float64
}

func (v Vector2) Add(o Vector2) Vector2 {
	return Vector2{v.X + o.X, v.Y + o.Y}
}

func (v Vector2) Sub(o Vector2) Vector2 {
	return Vector2{v.X - o.X, v.Y - o.Y}
}

func (v Vector2) Scale(s float64) Vector2 {
	return Vector2{v.X * s, v.Y * s}
}

func (v Vector2) Dot(o Vector2) float64 {
	return v.X*o.X + v.Y*o.Y
}

func (v Vector2) Length() float64 {
	return math.Hypot(v.X, v.Y)
}

func (v Vector2) Normalized() Vector2 {
	var length = v.Length()
	if length == 0 {
		return Vector2{}
	}

	return Vector2{v.X / length, v.Y / length}
}

func (v Vector2) DistanceTo(o Vector2) float64 {
	return o.Sub(v).Length()
}
