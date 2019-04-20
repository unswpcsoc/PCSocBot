from commands.base import Command

from random import choice, randrange


class Scream(Command):
    desc = "fnshkdhnjksdg"

    def eval(self):
        # roll the 50-sided dice
        roll = randrange(50)
        out = ""

        # thanks tiff for these
        if roll == 1:
            # rare change of HHHhhhhh
            out = trail_caps('a')[::-1]
        elif roll == 2:
            # waaaaAAAAAAAAA
            out = 'w' + trail_caps('a')
        elif roll == 3:
            # AAAAAAAAAAaaaaa
            out = trail_caps('a')[::-1]

        elif roll == 4:
            # eeeeeeeeEEEEEEE
            out = trail_caps('e')

        else:
            # regular homerow mashing
            # tends to use home row + b and n
            # screaming likely begins at the index/middle finger except for d
            screams = "asdfghjklbn"
            starts = "fjk"

            if randrange(2) == 1:
                screams = screams.upper()
                starts = starts.upper()

            out = choice(starts)
            for _ in range(randrange(5, 15)):
                c = choice(screams)
                # make sure we don't repeat a char unless it's a h
                if c != 'h' and c != 'H':
                    while c == out[-1]:
                        c = choice(screams)
                out += c

        return out


def trail_caps(char):
    return char * randrange(1, 10) + char.upper * randrange(5)
