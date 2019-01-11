from commands.base import Command
from helpers import code

import subprocess

GIT_BRANCH = r"git branch | grep \* | sed -rn 's/\*[ ]*(.*)/\1/p'"


class Branch(Command):
    desc = 'Returns what feature branch PCSocBot is currently running on.'

    def eval(self):
        return "Running on " + code(subprocess.getoutput(GIT_BRANCH))
