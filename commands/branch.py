from commands.base import Command
from helpers import code

import re, subprocess


class Branch(Command):
    desc = 'Returns what feature branch PCSocBot is currently running on.'

    def eval(self):
        output = subprocess.getoutput("git branch")
        branch = re.search('\n\* (.*?)\n', output).group(1)
        return "Running on " + code(branch)
