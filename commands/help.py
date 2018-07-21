from commands.base import Command
from helpers import *


class Helpme(Command):
    desc = 'This help text'
    pprint = dict(args='[<command/subcommand>]')

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
            return " ".join(args) + " is not a command.\n" + Command.help
        else:
            return cls.help + '\n\nType ' + Helpme.tag_markup + ' for more info on a command'

class H(Command):
    desc = "See " + bold(code("!helpme")) + "."

    def eval(args): return Helpme.eval(args)
