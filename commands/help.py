from commands.base import Command
from helpers import *


class Helpme(Command):
    desc = 'This help text'

    @staticmethod
    def find_command(args):
        # This function is too powerful to be a staticmethod in this class
        # It needs to be separate from commands so that config can run it safely
        cls = Command
        first_arg = 0
        
        while first_arg < len(args) and args[first_arg] in cls.subcommands:
            # Traverse user input until we find the exact command to use.
            # Navigate the command hierarchy to find the final subcommand
            # eg: !tags platforms == Command --> tags --> platforms
            cls = cls.subcommands[args[first_arg]]
            first_arg += 1

        # Return the (sub)Command and all its arguments
        return cls, args[first_arg:]

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
