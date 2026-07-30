extends Node2D

const ARENA_SIZE := Vector2(1152, 648)
const GRID_STEP := 64.0


func _draw() -> void:
	var field := Rect2(Vector2.ZERO, ARENA_SIZE)

	# Play-field floor, distinct from the darker letterbox bars around it.
	draw_rect(field, Palette.FLOOR, true)

	# Grid: thin reference lines, brighter every 4th line.
	var x := GRID_STEP
	var col := 1
	while x < ARENA_SIZE.x:
		var color := Palette.GRID_MAJOR if col % 4 == 0 else Palette.GRID
		draw_line(Vector2(x, 0), Vector2(x, ARENA_SIZE.y), color, 1.0)
		x += GRID_STEP
		col += 1

	var y := GRID_STEP
	var row := 1
	while y < ARENA_SIZE.y:
		var color := Palette.GRID_MAJOR if row % 4 == 0 else Palette.GRID
		draw_line(Vector2(0, y), Vector2(ARENA_SIZE.x, y), color, 1.0)
		y += GRID_STEP
		row += 1

	# Glowing border: wide translucent passes behind a bright thin edge.
	draw_rect(field, Palette.BORDER_GLOW, false, 8.0)
	draw_rect(field, Palette.BORDER_GLOW, false, 4.0)
	draw_rect(field, Palette.BORDER_STRONG, false, 2.0)
