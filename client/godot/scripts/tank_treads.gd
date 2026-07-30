extends Node2D

const TREAD_LEFT := -15.0
const TREAD_LENGTH := 30.0
const LINK_SPACING := 3.0  # Must divide TREAD_LENGTH so the wrap is seamless
const LINK_THICKNESS := 1.6
const LINK_INSET := 0.6  # Shrink links off the tread edges

# The tank faces +X; the treads run along X on each side. [top_y, bottom_y].
const BANDS := [Vector2(-13.0, -8.0), Vector2(8.0, 13.0)]

const TREAD_COLOR := Color(0.2, 0.21, 0.24)
const LINK_COLOR := Color(0.42, 0.44, 0.48)

var scroll := 0.0


func _draw() -> void:
	var link_count := int(TREAD_LENGTH / LINK_SPACING)
	for band in BANDS:
		var top: float = band.x
		var height: float = band.y - band.x

		# The dark tread the links ride on.
		draw_rect(Rect2(TREAD_LEFT, top, TREAD_LENGTH, height), TREAD_COLOR, true)

		# Evenly spaced links, shifted by `scroll` and wrapped with fposmod so
		# they recycle seamlessly from one end of the tread to the other.
		for i in link_count:
			var x := TREAD_LEFT + fposmod(i * LINK_SPACING - scroll, TREAD_LENGTH)
			var link := Rect2(
				x - LINK_THICKNESS * 0.5,
				top + LINK_INSET,
				LINK_THICKNESS,
				height - LINK_INSET * 2.0
			)
			draw_rect(link, LINK_COLOR, true)
