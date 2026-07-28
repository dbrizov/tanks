from datetime import datetime, timedelta, timezone

import jwt


def generate_token(username, secret, expiration_minutes):
    now = datetime.now(timezone.utc)
    expiry = now + timedelta(minutes=expiration_minutes)
    payload = {
        "sub": username,
        "iat": now,
        "exp": expiry,
    }

    return jwt.encode(payload, secret, algorithm="HS256")
