from flask import Blueprint, jsonify, request

from . import services
from .errors import AuthError

health_bp = Blueprint("health", __name__)
auth_bp = Blueprint("auth", __name__)

MIN_USERNAME_LENGTH = 3
MAX_USERNAME_LENGTH = 20
MIN_PASSWORD_LENGTH = 6
MAX_PASSWORD_LENGTH = 72


@health_bp.get("/health")
def health():
    return jsonify({"status": "ok"})


@auth_bp.post("/register")
def register():
    username, password = _credentials()
    _validate_username(username)
    _validate_password(password)
    services.register(username, password)
    return "", 201


@auth_bp.post("/login")
def login():
    username, password = _credentials()
    if not _present(username) or not _present(password):
        raise AuthError(400, "username and password are required")

    token = services.login(username, password)
    return jsonify({"token": token})


def _credentials():
    data = request.get_json(silent=True) or {}
    return data.get("username"), data.get("password")


def _present(value):
    return isinstance(value, str) and value.strip() != ""


def _validate_username(username):
    if not _present(username):
        raise AuthError(400, "username must not be blank")
    if not (MIN_USERNAME_LENGTH <= len(username) <= MAX_USERNAME_LENGTH):
        raise AuthError(400, f"username must be between {MIN_USERNAME_LENGTH} and {MAX_USERNAME_LENGTH} characters")


def _validate_password(password):
    if not _present(password):
        raise AuthError(400, "password must not be blank")
    if not (MIN_PASSWORD_LENGTH <= len(password) <= MAX_PASSWORD_LENGTH):
        raise AuthError(400, f"password must be between {MIN_PASSWORD_LENGTH} and {MAX_PASSWORD_LENGTH} characters")
