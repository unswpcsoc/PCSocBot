from pony.orm import Database, sql_debug

from helpers import CommandFailure
from configstartup import config

DB_FILE = config['FILES'].get('DB')

db = Database()
db.bind('sqlite', DB_FILE, create_db=True)


class Table(object):
    @classmethod
    def get_or_err(cls, err=None, **kwargs):
        obj = cls.get(**kwargs)
        if obj is None:
            raise CommandFailure(cls.err if err is None else err)
        return obj

    @classmethod
    def create_or_update(cls, **kwargs):
        old_obj = cls.get(**dict((k, kwargs[k]) for k in cls._pk_columns_))
        if old_obj is None:
            return cls(**kwargs)
        for k, v in kwargs.items():
            if k not in cls._pk_columns_:
                setattr(old_obj, k, v)
        return old_obj

    @classmethod
    def delete_or_err(cls, err=None, **kwargs):
        cls.get_or_err(err, **kwargs).delete()

    @classmethod
    def select_or_err(cls, fn, err=None):
        objs = cls.select(fn)
        if not objs:
            raise CommandFailure(cls.err if err is None else err)
        return objs

    @classmethod
    def select_fields_or_err(cls, fields, fn, err=None):
        return [[getattr(x, field) for field in fields] for x in cls.select_or_err(fn, err)]
