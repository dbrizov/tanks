extends Node

var token := ""
var username := ""
var auth_url := ""
var server_url := ""


func _ready() -> void:
	if OS.is_debug_build():
		auth_url = "http://127.0.0.1:8100"
		server_url = "ws://127.0.0.1:8101/play"
	else:
		auth_url = "https://dev.denisrizov.com/tanks/auth"
		server_url = "wss://dev.denisrizov.com/tanks/server/play"


func play_url() -> String:
	return "%s?token=%s" % [server_url, token.uri_encode()]


func clear() -> void:
	token = ""
	username = ""
