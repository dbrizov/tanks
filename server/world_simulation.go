package main

import (
	"math"
	"math/rand"
)

func (w *World) spawnTank(id PlayerId) *Tank {
	var pos = w.randomSpawnPos()
	return &Tank{
		Id:   id,
		PosX: pos.X,
		PosY: pos.Y,
		Hp:   w.config.TankMaxHp,
	}
}

func (w *World) respawnTank(tank *Tank) {
	var pos = w.randomSpawnPos()
	tank.PosX = pos.X
	tank.PosY = pos.Y
	tank.Hp = w.config.TankMaxHp
	tank.respawnTimer = 0
}

func (w *World) randomSpawnPos() Vector2 {
	var margin = math.Hypot(w.config.TankHalfWidth, w.config.TankHalfHeight)
	var posX = margin + rand.Float64()*(w.config.ArenaWidth-2*margin)
	var posY = margin + rand.Float64()*(w.config.ArenaHeight-2*margin)
	return Vector2{X: posX, Y: posY}
}

func (w *World) step(deltaTime float64) {
	for _, tank := range w.tanks {
		tank.fireCooldown -= deltaTime

		if tank.Hp <= 0 {
			tank.respawnTimer -= deltaTime
			if tank.respawnTimer <= 0 {
				w.respawnTank(tank)
			}

			continue
		}

		var inputs = w.currentInputs[tank.Id]

		var direction = Vector2{X: inputs.AxisX, Y: inputs.AxisY}.Normalized()
		var deltaPos = direction.Scale(w.config.TankSpeed * deltaTime)
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
			tank.fireCooldown = w.config.TankFireCooldown
			tank.Shots++
		}
	}

	w.advanceProjectiles(deltaTime)
	w.resolveCollisions()
	w.clampTanks()
}

func (w *World) fireProjectile(tank *Tank) {
	var direction = Vector2{X: math.Cos(tank.RotAim), Y: math.Sin(tank.RotAim)}
	var position = Vector2{X: tank.PosX, Y: tank.PosY}.Add(direction.Scale(w.config.TankBarrelLength))

	w.nextProjectileId++
	w.projectiles[w.nextProjectileId] = &Projectile{
		Id:    w.nextProjectileId,
		PosX:  position.X,
		PosY:  position.Y,
		velX:  direction.X * w.config.ProjectileSpeed,
		velY:  direction.Y * w.config.ProjectileSpeed,
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
			victim.Hp -= w.config.ProjectileDamage
			delete(w.projectiles, id)

			if victim.Hp <= 0 {
				victim.respawnTimer = w.config.TankRespawnDelay
				var killer = w.tanks[projectile.owner]
				if killer != nil {
					killer.Score++
				}
			}

			continue
		}

		if to.X < 0 || to.X > w.config.ArenaWidth || to.Y < 0 || to.Y > w.config.ArenaHeight {
			delete(w.projectiles, id)
		}
	}
}

func (w *World) projectileHit(projectile *Projectile, from Vector2, to Vector2) *Tank {
	for _, tank := range w.tanks {
		if tank.Id == projectile.owner || tank.Hp <= 0 {
			continue
		}

		if w.tankOBB(tank).IntersectsLineSegment(from, to, w.config.ProjectileRadius) {
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

			var separation, overlapping = w.tankOBB(tankA).Overlap(w.tankOBB(tankB))
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
		var extentX = w.config.TankHalfWidth*absCos + w.config.TankHalfHeight*absSin
		var extentY = w.config.TankHalfWidth*absSin + w.config.TankHalfHeight*absCos
		tank.PosX = clamp(tank.PosX, extentX, w.config.ArenaWidth-extentX)
		tank.PosY = clamp(tank.PosY, extentY, w.config.ArenaHeight-extentY)
	}
}

// tankOBB builds the oriented bounding box for a tank's body.
func (w *World) tankOBB(tank *Tank) OBB {
	return OBB{
		Center:  Vector2{X: tank.PosX, Y: tank.PosY},
		Extents: Vector2{X: w.config.TankHalfWidth, Y: w.config.TankHalfHeight},
		Rot:     tank.RotBody,
	}
}

func clamp(value float64, low float64, high float64) float64 {
	return min(max(value, low), high)
}
