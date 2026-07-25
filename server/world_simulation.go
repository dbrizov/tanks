package main

import (
	"math"
	"math/rand"
)

const (
	arenaWidth  = 800.0
	arenaHeight = 600.0
	tankRadius  = 15.0
	tankSpeed   = 200.0 // units per second
)

func spawnTank(id PlayerId) *TankState {
	var tankDiameter = tankRadius * 2
	var xOffset = rand.Float64() * (arenaWidth - tankDiameter)
	var yOffset = rand.Float64() * (arenaHeight - tankDiameter)

	return &TankState{
		Id: id,
		X:  tankRadius + xOffset,
		Y:  tankRadius + yOffset,
	}
}

func (w *World) step(delta_time float64) {
	for id, in := range w.currentInputs {
		var tank = w.tanks[id]
		if tank == nil {
			continue
		}

		var direction = Vec2{X: in.Ax, Y: in.Ay}.Normalized()
		var deltaPos = direction.Scale(tankSpeed * delta_time)
		tank.X += deltaPos.X
		tank.Y += deltaPos.Y

		if direction.X != 0 || direction.Y != 0 {
			tank.Angle = math.Atan2(direction.Y, direction.X)
		}
	}

	w.resolveCollisions()

	for _, tank := range w.tanks {
		var maxX = arenaWidth - tankRadius
		var maxY = arenaHeight - tankRadius
		tank.X = clamp(tank.X, tankRadius, maxX)
		tank.Y = clamp(tank.Y, tankRadius, maxY)
	}
}

func (w *World) resolveCollisions() {
	var tanks = make([]*TankState, 0, len(w.tanks))
	for _, t := range w.tanks {
		tanks = append(tanks, t)
	}

	const minDist = tankRadius * 2

	for i := 0; i < len(tanks); i++ {
		for j := i + 1; j < len(tanks); j++ {
			var a = tanks[i]
			var b = tanks[j]
			var aPos = Vec2{X: a.X, Y: a.Y}
			var bPos = Vec2{X: b.X, Y: b.Y}

			var dist = aPos.DistanceTo(bPos)
			if dist == 0 || dist >= minDist {
				continue
			}

			var push = bPos.Sub(aPos).Normalized().Scale((minDist - dist) / 2)
			a.X -= push.X
			a.Y -= push.Y
			b.X += push.X
			b.Y += push.Y
		}
	}
}

func clamp(v float64, lo float64, hi float64) float64 {
	return min(max(v, lo), hi)
}
