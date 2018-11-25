from commands.base import Command
# from helpers import *

class PigLatin(Command):
    desc = "This command is used to translate a given sentence into Pig Latin."

    def eval(self, *args):
        if args:
            return ' '.join([arg + 'ay' if arg[0].lower() in 'aeiou' else arg[1:] + arg[0] + 'ay' for arg in args])