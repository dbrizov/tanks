class_name PlayerColor
extends RefCounted


static func for_id(id: String) -> Color:
	var hue := float(absi(id.hash()) % 360) / 360.0
	return Color.from_hsv(hue, 0.6, 1.0)
