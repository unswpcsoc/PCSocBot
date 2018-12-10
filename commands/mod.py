import random

from commands.base import Command
from helpers import *


class Mod(Command):
    desc = "This command is used to mention a random mod in times of need. "
    "Not a smart idea to abuse this command"

    def eval(self): return "Paging %s..." % random.choice([x for x in list(
        self.members) if "mod" in [str(y).lower() for y in x.roles]]).mention
