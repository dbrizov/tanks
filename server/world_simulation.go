package main

import (
	"math"
	"math/rand"
)

const ticksPerSecond = 30

const arenaWidth = 1152.0
const arenaHeight = 648.0

const tankHalfWidth = 15.0
const tankHalfHeight = 13.0

const tankSpeed = 200.0 // units per second

func spawnTank(id PlayerId) *TankState {
	var margin = math.Hypot(tankHalfWidth, tankHalfHeight)
	var xOffset = rand.Float64() * (arenaWidth - 2*margin)
	var yOffset = rand.Float64() * (arenaHeight - 2*margin)

	return &TankState{
		Id:   id,
		PosX: margin + xOffset,
		PosY: margin + yOffset,
	}
}

func (w *World) step(delta_time float64) {
	for id, in := range w.currentInputs {
		var tank = w.tanks[id]
		if tank == nil {
			continue
		}

		var direction = Vector2{X: in.AxisX, Y: in.AxisY}.Normalized()
		var deltaPos = direction.Scale(tankSpeed * delta_time)
		tank.PosX += deltaPos.X
		tank.PosY += deltaPos.Y

		if direction.X != 0 || direction.Y != 0 {
			tank.RotBody = math.Atan2(direction.Y, direction.X)
		}

		var aimDirX = in.MouseX - tank.PosX
		var aimDirY = in.MouseY - tank.PosY
		tank.RotAim = math.Atan2(aimDirY, aimDirX)
	}

	w.resolveCollisions()

	for _, tank := range w.tanks {
		// A rotated box reaches further along the world axes than its half-extents.
		var c = math.Abs(math.Cos(tank.RotBody))
		var s = math.Abs(math.Sin(tank.RotBody))
		var ex = tankHalfWidth*c + tankHalfHeight*s
		var ey = tankHalfWidth*s + tankHalfHeight*c
		tank.PosX = clamp(tank.PosX, ex, arenaWidth-ex)
		tank.PosY = clamp(tank.PosY, ey, arenaHeight-ey)
	}
}

func (w *World) resolveCollisions() {
	var tanks = make([]*TankState, 0, len(w.tanks))
	for _, t := range w.tanks {
		tanks = append(tanks, t)
	}

	for i := 0; i < len(tanks); i++ {
		for j := i + 1; j < len(tanks); j++ {
			var a = tanks[i]
			var b = tanks[j]

			var separation, overlapping = tankOBB(a).Overlap(tankOBB(b))
			if !overlapping {
				continue
			}

			// Split the correction: push each tank half of the penetration apart.
			var push = separation.Scale(0.5)
			a.PosX -= push.X
			a.PosY -= push.Y
			b.PosX += push.X
			b.PosY += push.Y
		}
	}
}

// tankOBB builds the oriented bounding box for a tank's body.
func tankOBB(t *TankState) OBB {
	return OBB{
		Center:     Vector2{X: t.PosX, Y: t.PosY},
		HalfWidth:  tankHalfWidth,
		HalfHeight: tankHalfHeight,
		Rot:        t.RotBody,
	}
}

func clamp(v float64, lo float64, hi float64) float64 {
	return min(max(v, lo), hi)
}
