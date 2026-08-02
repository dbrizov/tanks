from flask import current_app

from .errors import AuthError
from .db import db
from .jwt_service import generate_token
from .models import User
from .security import hash_password, verify_password


def register(username, raw_password):
    existing = db.session.execute(
        db.select(User).filter_by(username=username)
    ).scalar_one_or_none()

    if existing is not None:
        raise AuthError(409, "username already taken")

    user = User(username, hash_password(raw_password))
    db.session.add(user)
    db.session.commit()


def login(username, raw_password):
    user = db.session.execute(
        db.select(User).filter_by(username=username)
    ).scalar_one_or_none()

    if user is None or not verify_password(raw_password, user.password_hash):
        raise AuthError(401, "invalid credentials")

    secret = current_app.config["JWT_SECRET"]
    expiration_minutes = current_app.config["JWT_EXPIRATION_MINUTES"]
    token = generate_token(username, secret, expiration_minutes)
    return token
