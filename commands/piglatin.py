from commands.base import Command
# from helpers import *

class PigLatin(Command):
    desc = "This command is used to translate a given sentence into Pig Latin."

    def eval(self, *args):
        suffix = 'ay'
        return ' '.join([arg + suffix if arg[0].lower() in 'aeiou' else arg[1:] + arg[0] + suffix for arg in args])