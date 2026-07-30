extends Node2D

const TRAIL_POINTS := 12

@onready var _trail: Line2D = $Trail


func _process(_delta: float) -> void:
	_trail.add_point(global_position)
	if _trail.get_point_count() > TRAIL_POINTS:
		_trail.remove_point(0)
