import os

from flask import Flask
from flask_cors import CORS

from .config import DevConfig, ProdConfig
from .errors import register_error_handlers
from .db import db
from .routes import auth_bp, health_bp


def create_app():
    app = Flask(__name__)

    env_name = os.environ.get("APP_ENV", "dev").lower()
    config = ProdConfig if env_name == "prod" else DevConfig
    app.config.from_object(config)

    if config.ENV_NAME == "prod" and not app.config.get("JWT_SECRET"):
        raise RuntimeError("JWT_SECRET must be set in the prod profile")

    db.init_app(app)

    CORS(
        app,
        origins=[o.strip() for o in app.config["CORS_ORIGINS"].split(",") if o.strip()],
        methods=["GET", "POST", "OPTIONS"],
    )

    app.register_blueprint(health_bp)
    app.register_blueprint(auth_bp)
    register_error_handlers(app)

    with app.app_context():
        db.create_all()

    return app
