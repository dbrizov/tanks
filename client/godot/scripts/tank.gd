extends Node2D

func set_state(pos: Vector2, rot_body: float, rot_aim: float) -> void:
	position = pos
	$Body.rotation = rot_body
	$Barrel.rotation = rot_aim
