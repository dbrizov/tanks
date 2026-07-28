from flask import Blueprint, jsonify, request

from . import services
from .errors import AuthError

health_bp = Blueprint("health", __name__)
auth_bp = Blueprint("auth", __name__)


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
    if not (3 <= len(username) <= 50):
        raise AuthError(400, "username must be between 3 and 50 characters")


def _validate_password(password):
    if not _present(password):
        raise AuthError(400, "password must not be blank")
    if not (6 <= len(password) <= 100):
        raise AuthError(400, "password must be between 6 and 100 characters")
