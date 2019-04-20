from commands.base import Command

from random import choice, randrange


class Scream(Command):
    desc = "fnshkdhnjksdg"

    def eval(self):
        # roll the 50-sided dice
        roll = randrange(0, 50)
        out = ""

        # rare change of HHHhhhhh
        if roll == 1:
            out = 'H' * randrange(0, 5) + 'h' * randrange(1, 10)

        # thanks tiff for these
        elif roll == 2:
            out = 'w' + 'a' * randrange(0, 5) + 'A' * randrange(1, 10)

        elif roll == 3:
            out = 'A' * randrange(1, 10) + 'a' * randrange(0, 5)

        elif roll == 4:
            out = 'E' * randrange(5, 15)

        return out

        # regular homerow mashing
        # tends to use home row + b and n
        screams = list("asdfghjklbn")
        # screaming likely begins at the index/middle finger except for d
        starts = list("fjk")

        if randrange(0, 2) == 1:
            screams = list(map(str.upper, screams))
            starts = list(map(str.upper, starts))
        else:
            screams = list(map(str.lower, screams))
            starts = list(map(str.lower, starts))
        out = choice(starts)

        for _ in range(randrange(5, 15)):
            c = choice(screams)
            # make sure we don't repeat a char unless it's a h
            if c != 'h' and c != 'H':
                while c == out[-1]:
                    c = choice(screams)
            out += c

        return out
