extends Node

signal config_loaded

var token := ""
var username := ""
var auth_url := "http://localhost:8081"
var server_url := "ws://localhost:8080/play"
var config_ready := false

var _http: HTTPRequest


func _ready() -> void:
	_http = HTTPRequest.new()
	add_child(_http)
	_http.timeout = 5.0
	_http.request_completed.connect(_on_config_loaded)
	var err := _http.request("config.json")
	if err != OK:
		_finish_config()


func _on_config_loaded(
	result: int, code: int, _headers: PackedStringArray, body: PackedByteArray
) -> void:
	if result == HTTPRequest.RESULT_SUCCESS and code == 200:
		var data = JSON.parse_string(body.get_string_from_utf8())
		if typeof(data) == TYPE_DICTIONARY:
			auth_url = str(data.get("auth_url", auth_url))
			server_url = str(data.get("server_url", server_url))
	_finish_config()


func _finish_config() -> void:
	if config_ready:
		return
	config_ready = true
	config_loaded.emit()


func play_url() -> String:
	return "%s?token=%s" % [server_url, token.uri_encode()]


func clear() -> void:
	token = ""
	username = ""
