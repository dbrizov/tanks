package main

import "math"

type Vector2 struct {
	X float64
	Y float64
}

func (v Vector2) Add(other Vector2) Vector2 {
	return Vector2{v.X + other.X, v.Y + other.Y}
}

func (v Vector2) Sub(other Vector2) Vector2 {
	return Vector2{v.X - other.X, v.Y - other.Y}
}

func (v Vector2) Scale(scalar float64) Vector2 {
	return Vector2{v.X * scalar, v.Y * scalar}
}

func (v Vector2) Dot(other Vector2) float64 {
	return v.X*other.X + v.Y*other.Y
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

func (v Vector2) DistanceTo(other Vector2) float64 {
	return other.Sub(v).Length()
}
