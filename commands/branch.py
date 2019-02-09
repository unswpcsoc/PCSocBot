from commands.base import Command
from helpers import code

import re
import subprocess

BRANCH = code(subprocess.getoutput(r'git rev-parse --abbrev-ref HEAD'))


class Branch(Command):
    desc = 'Returns what feature branch PCSocBot is currently running on.'

    def eval(self):
        return "Running on " + BRANCH
