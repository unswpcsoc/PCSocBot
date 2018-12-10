import json
import random

from commands.base import Command
from helpers import *

WISH_FILE = "files/wish.json"

class Wish(Command):
    desc = "Wish upon a star!"

    async def eval(self, *args):
        # Retrieve the wish list
        try:
            # Get the format dict, throws FileNotFoundError
            with open(WISH_FILE, 'r') as fmt:
                wishes = json.load(fmt)
        except FileNotFoundError:
            raise CommandFailure('Wish list is empty!')

        return random.choice(wishes)