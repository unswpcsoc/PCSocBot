from models.tags import *
from models.database import db, DB_FILE

db.generate_mapping(create_tables=True, check_tables=True)
