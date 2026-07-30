extends Node2D

const MAX_HP := 100.0

const SMOOTH_POS := 18.0
const SMOOTH_BODY := 12.0
const SMOOTH_AIM := 22.0
const FADE_SPEED := 6.0
const TREAD_MOVE_EPSILON := 0.4  # Below this per-frame travel the treads hold still

const HP_COLOR_SELF := Color(0.2, 0.85, 0.3)
const HP_COLOR_ENEMY := Color(0.9, 0.25, 0.25)

var _target_pos := Vector2.ZERO
var _target_body_rot := 0.0
var _target_aim_rot := 0.0
var _hp := int(MAX_HP)
var _alive := true
var _target_alpha := 1.0
var _tread_offset := 0.0
var _shot_count := -1  # -1 = uninitialized; first snapshot sets it without flashing
var _initialized := false

@onready var _visual: Node2D = $Visual
@onready var _hull: Node2D = $Visual/Hull
@onready var _turret: Node2D = $Visual/Turret
@onready var _body_poly: Polygon2D = $Visual/Hull/Body
@onready var _turret_base: Polygon2D = $Visual/Turret/Base
@onready var _hp_fill: Polygon2D = $Visual/HpBarFill
@onready var _name_label: Label = $Visual/Name
@onready var _muzzle: CPUParticles2D = $Visual/Turret/Muzzle
@onready var _explosion: CPUParticles2D = $Explosion
@onready var _treads = $Visual/Hull/Treads  # tank_treads.gd; untyped for dynamic `scroll`


func set_label(text: String) -> void:
	_name_label.text = text


func set_identity(id: String, is_self: bool) -> void:
	var color := PlayerColor.for_id(id)
	_body_poly.color = color
	_turret_base.color = color.darkened(0.3)
	_hp_fill.color = HP_COLOR_SELF if is_self else HP_COLOR_ENEMY


func set_state(pos: Vector2, rot_body: float, rot_aim: float, hp: int) -> void:
	_target_pos = pos
	_target_body_rot = rot_body
	_target_aim_rot = rot_aim
	_hp = hp
	var now_alive := hp > 0
	var just_died := _alive and not now_alive
	var respawned := not _alive and now_alive

	if not _initialized:
		_initialized = true
		_snap_to(pos, rot_body, rot_aim)
		_alive = now_alive
		_target_alpha = 1.0 if now_alive else 0.0
		_visual.modulate.a = _target_alpha
	elif just_died:
		_explosion.restart()
		_explosion.emitting = true
		_target_alpha = 0.0
	elif respawned:
		_snap_to(pos, rot_body, rot_aim)
		_target_alpha = 1.0

	_alive = now_alive
	_update_hp_bar()


func update_shots(count: int) -> void:
	if _shot_count >= 0 and count > _shot_count:
		flash_muzzle()
	_shot_count = count


func flash_muzzle() -> void:
	_muzzle.restart()
	_muzzle.emitting = true


func _process(delta: float) -> void:
	if not _initialized:
		return

	var new_pos := position.lerp(_target_pos, 1.0 - exp(-SMOOTH_POS * delta))
	var motion := new_pos - position
	position = new_pos

	# Scroll the treads by the distance travelled this frame
	var moved := motion.length()
	if moved > TREAD_MOVE_EPSILON:
		var facing := Vector2.RIGHT.rotated(_hull.rotation)
		var forward := 1.0 if facing.dot(motion) >= 0.0 else -1.0
		_tread_offset += forward * moved
		_treads.scroll = _tread_offset
		_treads.queue_redraw()

	_hull.rotation = lerp_angle(_hull.rotation, _target_body_rot, 1.0 - exp(-SMOOTH_BODY * delta))
	_turret.rotation = lerp_angle(_turret.rotation, _target_aim_rot, 1.0 - exp(-SMOOTH_AIM * delta))

	if not is_equal_approx(_visual.modulate.a, _target_alpha):
		_visual.modulate.a = move_toward(_visual.modulate.a, _target_alpha, FADE_SPEED * delta)


func _snap_to(pos: Vector2, rot_body: float, rot_aim: float) -> void:
	position = pos
	_hull.rotation = rot_body
	_turret.rotation = rot_aim


func _update_hp_bar() -> void:
	_hp_fill.scale.x = clampf(_hp / MAX_HP, 0.0, 1.0)
