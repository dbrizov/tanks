extends Node2D

const SEND_INTERVAL := 1.0 / 30.0
const RECONNECT_DELAY := 1.0

@export var tank_scene: PackedScene = preload("res://scenes/Tank.tscn")
@export var projectile_scene: PackedScene = preload("res://scenes/Projectile.tscn")

var _socket := WebSocketPeer.new()
var _tanks := {}  # PlayerId (String) -> Tank node
var _projectiles := {}  # Projectile id (String) -> Projectile node
var _self_id := ""
var _send_accum := 0.0
var _reconnect_accum := 0.0

@onready var _tanks_root: Node2D = $Tanks
@onready var _projectiles_root: Node2D = $Projectiles
@onready var _status: Label = $HUD/Status


func _ready() -> void:
	if Session.token.is_empty():
		get_tree().change_scene_to_file("res://scenes/Login.tscn")
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

		var pos: Vector2 = Vector2(tank_data.x, tank_data.y)
		var rot_body: float = tank_data.rb
		var rot_aim: float = tank_data.ra
		var hp: int = tank_data.hp
		tank.set_state(pos, rot_body, rot_aim, hp)

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

			var node: Node2D = _projectiles.get(id)
			if node == null:
				node = projectile_scene.instantiate()
				_projectiles_root.add_child(node)
				_projectiles[id] = node

			node.position = Vector2(proj_data.x, proj_data.y)

	for existing_id in _projectiles.keys():
		if not present.has(existing_id):
			_projectiles[existing_id].queue_free()
			_projectiles.erase(existing_id)


func _update_scoreboard(snapshot: Dictionary) -> void:
	var scores := {}  # id (String) -> score
	for tank_data in snapshot.tanks:
		scores[str(tank_data.id)] = int(tank_data.sc)

	var ids: Array = scores.keys()
	ids.sort_custom(
		func(a_id, b_id):
			var a_score: int = scores[a_id]
			var b_score: int = scores[b_id]
			if a_score != b_score:
				return a_score > b_score
			return a_id < b_id  # stable tiebreak by id so equal scores don't reshuffle
	)

	var lines := ["Connected | Tanks %d" % ids.size()]
	for id in ids:
		var who: String = str(id)
		lines.append("%s: %d" % [who, scores[id]])

	_status.text = "\n".join(lines)


func _reset() -> void:
	_self_id = ""
	for tank in _tanks.values():
		tank.queue_free()
	_tanks.clear()

	for node in _projectiles.values():
		node.queue_free()
	_projectiles.clear()
