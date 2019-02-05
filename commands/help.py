from commands.base import Command
from helpers import *


class Helpme(Command):
    desc = 'This help text'

    @staticmethod
    def find_command(args):
        cls = Command
        first_arg = 0
        for arg in args:
            if arg in cls.subcommands:
                cls = cls.subcommands[arg]
                first_arg += 1
            else:
                break
        args = args[first_arg:]
        return cls, args

    def eval(self, *args):
        cls, fn_args = Helpme.find_command(args)
        if fn_args:
            out = []
            if len(args) > 0:
                out.append(" ".join(args) + " is not a command.")
            out.extend(Command.help)
        else:
            out = cls.help
        return out 

class H(Command):
    desc = "See " + bold(code("!helpme")) + "."

    def eval(self, *args): return Helpme.eval(self, *args)
