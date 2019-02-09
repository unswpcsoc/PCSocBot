from commands.base import Command


class PigLatin(Command):
    desc = "This command is used to translate a given sentence into Pig Latin."

    def eval(self, *args):
        suffix = 'ay'
        res = []
        if args:
            # There are arguments and it's not PCSocBot conducting the command
            for arg in args:
                if arg[0].lower() in 'aeiou':
                    # Starts with a vowel
                    res.append(arg + suffix)
                elif arg[0].isalpha():
                    # Consonant
                    res.append(arg[1:] + arg[0] + suffix)
                else:
                    # Anything else (eg emotes, mentions)
                    res.append(arg)
        return ' '.join(res) if res else None
