from commands.base import Command
from helpers import *
from configstartup import config

import random
import json

WISH_FILE = config['FILES'].get('Wish')


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
