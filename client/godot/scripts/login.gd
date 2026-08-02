extends Control

var _pending_login := false

@onready var _username: LineEdit = $Center/Panel/Margin/VBox/Username
@onready var _password: LineEdit = $Center/Panel/Margin/VBox/Password
@onready var _login_button: Button = $Center/Panel/Margin/VBox/Buttons/Login
@onready var _register_button: Button = $Center/Panel/Margin/VBox/Buttons/Register
@onready var _status: Label = $Center/Panel/Margin/VBox/Status
@onready var _http: HTTPRequest = $Http


func _ready() -> void:
	Session.clear()
	_http.request_completed.connect(_on_request_completed)
	_login_button.pressed.connect(_on_login_pressed)
	_register_button.pressed.connect(_on_register_pressed)
	_password.text_submitted.connect(func(_t): _on_login_pressed())


func _on_login_pressed() -> void:
	_submit("/login", true)


func _on_register_pressed() -> void:
	_submit("/register", false)


func _submit(path: String, is_login: bool) -> void:
	var username := _username.text.strip_edges()
	var password := _password.text
	if username.is_empty() or password.is_empty():
		_status.text = "Enter a username and password"
		return

	_pending_login = is_login
	_set_busy(true)
	_status.text = "Logging in…" if is_login else "Registering…"

	var body := JSON.stringify({"username": username, "password": password})
	var headers := ["Content-Type: application/json"]
	var err := _http.request(Session.auth_url + path, headers, HTTPClient.METHOD_POST, body)
	if err != OK:
		_set_busy(false)
		_status.text = "Request failed to start (err %d)" % err


func _on_request_completed(
	result: int, code: int, _headers: PackedStringArray, body: PackedByteArray
) -> void:
	_set_busy(false)

	if result != HTTPRequest.RESULT_SUCCESS:
		_status.text = "Cannot reach auth service"
		return

	if _pending_login:
		_handle_login_response(code, body)
	else:
		_handle_register_response(code, body)


func _handle_login_response(code: int, body: PackedByteArray) -> void:
	if code == 200:
		var data = JSON.parse_string(body.get_string_from_utf8())
		if typeof(data) == TYPE_DICTIONARY and data.has("token"):
			Session.token = str(data.token)
			Session.username = _username.text.strip_edges()
			get_tree().change_scene_to_file("res://scenes/main.tscn")
			return
		_status.text = "Unexpected response from server"
	elif code == 401 or code == 403:
		_status.text = "Invalid username or password"
	else:
		_status.text = _error_message(body, "Login failed (HTTP %d)" % code)


func _handle_register_response(code: int, body: PackedByteArray) -> void:
	if code == 201 or code == 200:
		_status.text = "Registered — now log in"
	elif code == 409:
		_status.text = _error_message(body, "Username already taken")
	else:
		_status.text = _error_message(body, "Register failed (HTTP %d)" % code)


func _error_message(body: PackedByteArray, fallback: String) -> String:
	var data = JSON.parse_string(body.get_string_from_utf8())
	if typeof(data) == TYPE_DICTIONARY and data.has("error"):
		return str(data.error)
	return fallback


func _set_busy(busy: bool) -> void:
	_login_button.disabled = busy
	_register_button.disabled = busy
