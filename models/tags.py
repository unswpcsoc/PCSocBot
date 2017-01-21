from pony.orm import Required, PrimaryKey

from models.database import Table, db


class Tag(db.Entity, Table):
    err = "Platform/Game not found"
    user = Required(int, size=64)
    platform = Required(str)
    PrimaryKey(user, platform)
    tag = Required(str)
