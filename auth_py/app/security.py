import bcrypt


def hash_password(raw_password):
    hashed = bcrypt.hashpw(raw_password.encode("utf-8"), bcrypt.gensalt())
    return hashed.decode("utf-8")


def verify_password(raw_password, password_hash):
    return bcrypt.checkpw(raw_password.encode("utf-8"), password_hash.encode("utf-8"))
