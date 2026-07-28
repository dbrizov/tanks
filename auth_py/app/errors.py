from flask import jsonify


class AuthError(Exception):
    def __init__(self, status_code, message):
        super().__init__(message)
        self.status_code = status_code
        self.message = message


def register_error_handlers(app):
    @app.errorhandler(AuthError)
    def handle_auth_error(error):
        return jsonify({"error": error.message}), error.status_code
