import os

BASE_DIR = os.path.abspath(os.path.dirname(os.path.dirname(__file__)))
DEV_JWT_SECRET = "some-very-very-long-random-string-at-least-32-bytes-long"


class Config:
    JWT_EXPIRATION_MINUTES = 120
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    MAX_CONTENT_LENGTH = 16 * 1024  # 16 KB


class DevConfig(Config):
    ENV_NAME = "dev"
    JWT_SECRET = os.environ.get("JWT_SECRET", DEV_JWT_SECRET)
    CORS_ORIGINS = os.environ.get("APP_CORS_ALLOWED_ORIGINS", r"http://localhost:\d+,http://127.0.0.1:\d+")
    DB_PATH = os.environ.get("DB_PATH", os.path.join(BASE_DIR, "auth.db"))
    SQLALCHEMY_DATABASE_URI = f"sqlite:///{DB_PATH}"


class ProdConfig(Config):
    ENV_NAME = "prod"
    JWT_SECRET = os.environ.get("JWT_SECRET")
    CORS_ORIGINS = os.environ.get("APP_CORS_ALLOWED_ORIGINS", "https://denisrizov.com")
    DB_PATH = os.environ.get("DB_PATH", "/var/lib/tanks/auth.db")
    SQLALCHEMY_DATABASE_URI = f"sqlite:///{DB_PATH}"
