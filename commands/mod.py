from commands.base import Command
from helpers import *
from configstartup import config


import random

MOD_ROLE = config['ROLES'].get('Mod')


class Mod(Command):
    desc = "This command is used to mention a random mod in times of need. Not a smart idea to abuse this command"

    def eval(self):
        return "Paging %s..." % random.choice([x for x in list(self.members) if MOD_ROLE in [y.id for y in x.roles]]).mention
