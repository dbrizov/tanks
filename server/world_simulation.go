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
const tankBarrelLength = 24.0

const tankSpeed = 200.0 // units per second
const tankMaxHp = 100

const tankFireCooldown = 0.35 // seconds between shots
const tankRespawnDelay = 2.0  // seconds a dead tank waits before respawning

const projectileSpeed = 600.0 // units per second
const projectileRadius = 4.0
const projectileDamage = 25

func spawnTank(id PlayerId) *Tank {
	var pos = randomSpawnPos()
	return &Tank{
		Id:   id,
		PosX: pos.X,
		PosY: pos.Y,
		Hp:   tankMaxHp,
	}
}

func respawnTank(tank *Tank) {
	var pos = randomSpawnPos()
	tank.PosX = pos.X
	tank.PosY = pos.Y
	tank.Hp = tankMaxHp
	tank.respawnTimer = 0
}

func randomSpawnPos() Vector2 {
	var margin = math.Hypot(tankHalfWidth, tankHalfHeight)
	var posX = margin + rand.Float64()*(arenaWidth-2*margin)
	var posY = margin + rand.Float64()*(arenaHeight-2*margin)
	return Vector2{X: posX, Y: posY}
}

func (w *World) step(deltaTime float64) {
	for _, tank := range w.tanks {
		tank.fireCooldown -= deltaTime

		if tank.Hp <= 0 {
			tank.respawnTimer -= deltaTime
			if tank.respawnTimer <= 0 {
				respawnTank(tank)
			}

			continue
		}

		var inputs = w.currentInputs[tank.Id]

		var direction = Vector2{X: inputs.AxisX, Y: inputs.AxisY}.Normalized()
		var deltaPos = direction.Scale(tankSpeed * deltaTime)
		tank.PosX += deltaPos.X
		tank.PosY += deltaPos.Y

		if direction.X != 0 || direction.Y != 0 {
			tank.RotBody = math.Atan2(direction.Y, direction.X)
		}

		var aimDirX = inputs.MouseX - tank.PosX
		var aimDirY = inputs.MouseY - tank.PosY
		tank.RotAim = math.Atan2(aimDirY, aimDirX)

		if inputs.Fire && tank.fireCooldown <= 0 {
			w.fireProjectile(tank)
			tank.fireCooldown = tankFireCooldown
		}
	}

	w.advanceProjectiles(deltaTime)
	w.resolveCollisions()
	w.clampTanks()
}

func (w *World) fireProjectile(tank *Tank) {
	var direction = Vector2{X: math.Cos(tank.RotAim), Y: math.Sin(tank.RotAim)}
	var position = Vector2{X: tank.PosX, Y: tank.PosY}.Add(direction.Scale(tankBarrelLength))

	w.nextProjectileId++
	w.projectiles[w.nextProjectileId] = &Projectile{
		Id:    w.nextProjectileId,
		PosX:  position.X,
		PosY:  position.Y,
		velX:  direction.X * projectileSpeed,
		velY:  direction.Y * projectileSpeed,
		owner: tank.Id,
	}
}

func (w *World) advanceProjectiles(deltaTime float64) {
	for id, projectile := range w.projectiles {
		var from = Vector2{X: projectile.PosX, Y: projectile.PosY}
		projectile.PosX += projectile.velX * deltaTime
		projectile.PosY += projectile.velY * deltaTime
		var to = Vector2{X: projectile.PosX, Y: projectile.PosY}

		var victim = w.projectileHit(projectile, from, to)
		if victim != nil {
			victim.Hp -= projectileDamage
			delete(w.projectiles, id)

			if victim.Hp <= 0 {
				victim.respawnTimer = tankRespawnDelay
				var killer = w.tanks[projectile.owner]
				if killer != nil {
					killer.Score++
				}
			}

			continue
		}

		if to.X < 0 || to.X > arenaWidth || to.Y < 0 || to.Y > arenaHeight {
			delete(w.projectiles, id)
		}
	}
}

func (w *World) projectileHit(projectile *Projectile, from Vector2, to Vector2) *Tank {
	for _, tank := range w.tanks {
		if tank.Id == projectile.owner || tank.Hp <= 0 {
			continue
		}

		if tankOBB(tank).IntersectsLineSegment(from, to, projectileRadius) {
			return tank
		}
	}

	return nil
}

func (w *World) resolveCollisions() {
	var tanks = make([]*Tank, 0, len(w.tanks))
	for _, tank := range w.tanks {
		if tank.Hp <= 0 {
			continue
		}

		tanks = append(tanks, tank)
	}

	for i := 0; i < len(tanks); i++ {
		for j := i + 1; j < len(tanks); j++ {
			var tankA = tanks[i]
			var tankB = tanks[j]

			var separation, overlapping = tankOBB(tankA).Overlap(tankOBB(tankB))
			if !overlapping {
				continue
			}

			// Split the correction: push each tank half of the penetration apart.
			var push = separation.Scale(0.5)
			tankA.PosX -= push.X
			tankA.PosY -= push.Y
			tankB.PosX += push.X
			tankB.PosY += push.Y
		}
	}
}

func (w *World) clampTanks() {
	for _, tank := range w.tanks {
		if tank.Hp <= 0 {
			continue
		}

		// A rotated box reaches further along the world axes than its half-extents.
		var absCos = math.Abs(math.Cos(tank.RotBody))
		var absSin = math.Abs(math.Sin(tank.RotBody))
		var extentX = tankHalfWidth*absCos + tankHalfHeight*absSin
		var extentY = tankHalfWidth*absSin + tankHalfHeight*absCos
		tank.PosX = clamp(tank.PosX, extentX, arenaWidth-extentX)
		tank.PosY = clamp(tank.PosY, extentY, arenaHeight-extentY)
	}
}

// tankOBB builds the oriented bounding box for a tank's body.
func tankOBB(tank *Tank) OBB {
	return OBB{
		Center:  Vector2{X: tank.PosX, Y: tank.PosY},
		Extents: Vector2{X: tankHalfWidth, Y: tankHalfHeight},
		Rot:     tank.RotBody,
	}
}

func clamp(value float64, low float64, high float64) float64 {
	return min(max(value, low), high)
}
