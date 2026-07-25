extends Node2D

const SEND_INTERVAL := 1.0 / 30.0
const RECONNECT_DELAY := 1.0

@export var tank_scene: PackedScene = preload("res://scenes/Tank.tscn")
@export var server_url := "ws://localhost:8080/play"

var _socket := WebSocketPeer.new()
var _tanks := {}  # PlayerId (String) -> Tank node
var _send_accum := 0.0
var _reconnect_accum := 0.0

@onready var _tanks_root: Node2D = $Tanks
@onready var _status: Label = $HUD/Status


func _ready() -> void:
	_socket.connect_to_url(server_url)


func _process(delta: float) -> void:
	_socket.poll()

	match _socket.get_ready_state():
		WebSocketPeer.STATE_OPEN:
			_reconnect_accum = 0.0
			_send_intent(delta)
			_drain_snapshots()

		WebSocketPeer.STATE_CLOSED:
			_status.text = "disconnected — retrying…"
			_reconnect_accum += delta
			if _reconnect_accum >= RECONNECT_DELAY:
				_reconnect_accum = 0.0
				_socket = WebSocketPeer.new()
				_socket.connect_to_url(server_url)

		_:
			_status.text = "connecting…"


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


func _drain_snapshots() -> void:
	var latest := ""
	while _socket.get_available_packet_count() > 0:
		latest = _socket.get_packet().get_string_from_utf8()

	if latest.is_empty():
		return

	var snapshot = JSON.parse_string(latest)
	if typeof(snapshot) != TYPE_DICTIONARY or not snapshot.has("tanks"):
		return

	_apply_snapshot(snapshot)


func _apply_snapshot(snapshot: Dictionary) -> void:
	var tanks_in_snapshot := {}

	for tank_data in snapshot.tanks:
		var id := str(tank_data.id)
		tanks_in_snapshot[id] = true

		var tank: Node2D = _tanks.get(id)
		if tank == null:
			tank = tank_scene.instantiate()
			_tanks_root.add_child(tank)
			_tanks[id] = tank

		var pos: Vector2 = Vector2(tank_data.x, tank_data.y)
		var rot_body: float = tank_data.rb
		var rot_aim: float = tank_data.ra
		tank.set_state(pos, rot_body, rot_aim)

	# Despawn tanks that are no longer in the snapshot.
	for existing_id in _tanks.keys():
		if not tanks_in_snapshot.has(existing_id):
			_tanks[existing_id].queue_free()
			_tanks.erase(existing_id)

	_status.text = "connected | tanks %d" % _tanks.size()
