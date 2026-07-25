extends Node2D

const MAX_HP := 100.0


func set_state(pos: Vector2, rot_body: float, rot_aim: float, hp: int) -> void:
	visible = hp > 0
	if hp <= 0:
		return

	position = pos
	$Body.rotation = rot_body
	$Barrel.rotation = rot_aim
	$HpBarFill.scale.x = clampf(hp / MAX_HP, 0.0, 1.0)
