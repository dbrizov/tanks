extends Node2D

const SEND_INTERVAL := 1.0 / 30.0
const RECONNECT_DELAY := 1.0

@export var tank_scene: PackedScene = preload("res://scenes/tank.tscn")
@export var projectile_scene: PackedScene = preload("res://scenes/projectile.tscn")
@export var hit_spark_scene: PackedScene = preload("res://scenes/hit_spark.tscn")

var _socket := WebSocketPeer.new()
var _tanks := {}  # PlayerId (String) -> Tank node
var _projectiles := {}  # Projectile id (String) -> Projectile node
var _self_id := ""
var _send_accum := 0.0
var _reconnect_accum := 0.0
var _scoreboard_sig := ""  # Last-rendered scoreboard, to skip rebuilding unchanged rows

@onready var _tanks_root: Node2D = $Tanks
@onready var _projectiles_root: Node2D = $Projectiles
@onready var _effects_root: Node2D = $Effects
@onready var _status: Label = $HUD/Status
@onready var _scores: GridContainer = $HUD/Scoreboard/Margin/VBox/Scores


func _ready() -> void:
	if Session.token.is_empty():
		get_tree().change_scene_to_file("res://scenes/login.tscn")
		return
	_socket.connect_to_url(Session.play_url())


func _process(delta: float) -> void:
	_socket.poll()

	match _socket.get_ready_state():
		WebSocketPeer.STATE_OPEN:
			_reconnect_accum = 0.0
			_send_intent(delta)
			_drain_packets()

		WebSocketPeer.STATE_CLOSED:
			_status.text = "Disconnected — retrying…"
			_reconnect_accum += delta
			if _reconnect_accum >= RECONNECT_DELAY:
				_reconnect_accum = 0.0
				_reset()
				_socket = WebSocketPeer.new()
				_socket.connect_to_url(Session.play_url())

		_:
			_status.text = "Connecting…"


func _send_intent(delta: float) -> void:
	_send_accum += delta
	if _send_accum < SEND_INTERVAL:
		return
	_send_accum = 0.0

	var mouse_pos := get_global_mouse_position()
	var intent := {
		"ax": Input.get_axis("left", "right"),  # -1 left, +1 right
		"ay": Input.get_axis("up", "down"),  # -1 up, +1 down (Y-down, matches server)
		"mx": mouse_pos.x,
		"my": mouse_pos.y,
		"fire": Input.is_action_pressed("fire"),
	}
	_socket.send_text(JSON.stringify(intent))


func _drain_packets() -> void:
	while _socket.get_available_packet_count() > 0:
		var text := _socket.get_packet().get_string_from_utf8()
		var msg = JSON.parse_string(text)
		if typeof(msg) != TYPE_DICTIONARY or not msg.has("type"):
			continue

		match msg.type:
			"joined":
				_self_id = str(msg.data)
			"snapshot":
				_apply_snapshot(msg.data)


func _apply_snapshot(snapshot: Dictionary) -> void:
	_apply_tanks(snapshot)
	_apply_projectiles(snapshot)
	_apply_impacts(snapshot)
	_update_status(snapshot)
	_update_scoreboard(snapshot)


func _apply_tanks(snapshot: Dictionary) -> void:
	var present := {}
	for tank_data in snapshot.tanks:
		var id := str(tank_data.id)
		present[id] = true

		var tank: Node2D = _tanks.get(id)
		if tank == null:
			tank = tank_scene.instantiate()
			_tanks_root.add_child(tank)
			_tanks[id] = tank
			tank.set_label(id)
			tank.set_identity(id, id == _self_id)

		var pos: Vector2 = Vector2(tank_data.x, tank_data.y)
		var rot_body: float = tank_data.rb
		var rot_aim: float = tank_data.ra
		var hp: int = tank_data.hp
		tank.set_state(pos, rot_body, rot_aim, hp)
		tank.update_shots(int(tank_data.get("sh", 0)))

	for existing_id in _tanks.keys():
		if not present.has(existing_id):
			_tanks[existing_id].queue_free()
			_tanks.erase(existing_id)


func _apply_projectiles(snapshot: Dictionary) -> void:
	var present := {}
	if snapshot.get("projectiles") != null:
		for proj_data in snapshot.projectiles:
			var id := str(proj_data.id)
			present[id] = true

			var pos: Vector2 = Vector2(proj_data.x, proj_data.y)
			var node: Node2D = _projectiles.get(id)
			if node == null:
				node = projectile_scene.instantiate()
				_projectiles_root.add_child(node)
				_projectiles[id] = node

			node.position = pos

	for existing_id in _projectiles.keys():
		if not present.has(existing_id):
			_projectiles[existing_id].queue_free()
			_projectiles.erase(existing_id)


func _apply_impacts(snapshot: Dictionary) -> void:
	var impacts = snapshot.get("impacts")
	if impacts == null:
		return

	for impact in impacts:
		_spawn_hit_spark(Vector2(impact.x, impact.y))


func _spawn_hit_spark(pos: Vector2) -> void:
	var spark := hit_spark_scene.instantiate()
	spark.position = pos
	_effects_root.add_child(spark)


func _update_status(snapshot: Dictionary) -> void:
	_status.text = "Connected · %d online" % snapshot.tanks.size()


func _update_scoreboard(snapshot: Dictionary) -> void:
	var scores := {}  # PlayerId (String) -> score
	for tank_data in snapshot.tanks:
		scores[str(tank_data.id)] = int(tank_data.sc)

	var ids: Array = scores.keys()
	ids.sort_custom(
		func(a_id, b_id):
			var a_score: int = scores[a_id]
			var b_score: int = scores[b_id]
			if a_score != b_score:
				return a_score > b_score
			return a_id < b_id  # Stable tiebreak by id so equal scores don't reshuffle
	)

	# Only rebuild the row nodes when the names/scores actually change, not every snapshot
	var sig := ""
	for id in ids:
		sig += "%s=%d;" % [str(id), scores[id]]
	if sig == _scoreboard_sig:
		return
	_scoreboard_sig = sig

	for cell in _scores.get_children():
		cell.free()

	for id in ids:
		var color: Color = PlayerColor.for_id(id)
		_scores.add_child(_make_score_cell(str(id), color, HORIZONTAL_ALIGNMENT_LEFT, false))
		_scores.add_child(
			_make_score_cell(str(scores[id]), Palette.TEXT, HORIZONTAL_ALIGNMENT_RIGHT, true)
		)


func _make_score_cell(
	text: String, color: Color, align: HorizontalAlignment, expand: bool
) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", color)
	label.horizontal_alignment = align
	if expand:
		# Let the score column absorb the slack so numbers sit flush right.
		label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	return label


func _reset() -> void:
	_self_id = ""
	_scoreboard_sig = ""

	for cell in _scores.get_children():
		cell.free()

	for tank in _tanks.values():
		tank.queue_free()
	_tanks.clear()

	for node in _projectiles.values():
		node.queue_free()
	_projectiles.clear()
