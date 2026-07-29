package main

import (
	"encoding/json"
	"log"
	"os"
)

const ConfigPath = "config.json"

type Config struct {
	TicksPerSecond int `json:"ticks_per_second"`

	ArenaWidth  float64 `json:"arena_width"`
	ArenaHeight float64 `json:"arena_height"`

	TankHalfWidth    float64 `json:"tank_half_width"`
	TankHalfHeight   float64 `json:"tank_half_height"`
	TankBarrelLength float64 `json:"tank_barrel_length"`
	TankSpeed        float64 `json:"tank_speed"` // units per second
	TankMaxHp        int     `json:"tank_max_hp"`
	TankFireCooldown float64 `json:"tank_fire_cooldown"` // seconds between shots
	TankRespawnDelay float64 `json:"tank_respawn_delay"` // seconds a dead tank waits before respawning

	ProjectileSpeed  float64 `json:"projectile_speed"` // units per second
	ProjectileRadius float64 `json:"projectile_radius"`
	ProjectileDamage int     `json:"projectile_damage"`
}

func defaultConfig() Config {
	return Config{
		TicksPerSecond: 30,

		ArenaWidth:  1152.0,
		ArenaHeight: 648.0,

		TankHalfWidth:    15.0,
		TankHalfHeight:   13.0,
		TankBarrelLength: 24.0,
		TankSpeed:        200.0,
		TankMaxHp:        100,

		TankFireCooldown: 0.35,
		TankRespawnDelay: 2.0,

		ProjectileSpeed:  600.0,
		ProjectileRadius: 4.0,
		ProjectileDamage: 25,
	}
}

func loadConfig() Config {
	var path = os.Getenv("CONFIG")
	if path == "" {
		path = ConfigPath
	}

	var data, err = os.ReadFile(path)
	if err != nil {
		log.Printf("config: using defaults (%v)", err)
		return defaultConfig()
	}

	var config = defaultConfig()
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("config: invalid %s, using defaults (%v)", path, err)
		return defaultConfig()
	}

	return config
}
