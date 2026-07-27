extends Node

var token := ""
var username := ""
var auth_url := "http://localhost:8081"
var server_url := "ws://localhost:8080/play"

func play_url() -> String:
	return "%s?token=%s" % [server_url, token.uri_encode()]


func clear() -> void:
	token = ""
	username = ""
