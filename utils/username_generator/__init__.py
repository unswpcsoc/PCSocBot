"""Username generator.

A command to generate a username from an int uid.
Modified from https://pypi.org/project/username-generator/

"""

from .wordlists import ADJECTIVES, NOUNS
import time
import random
max_colour = int('FFFFFF', 16)

time_salt = int(time.time())
random.seed(time_salt)
print("username_generator: time_salt is " + str(time_salt))
colours = dict()

def get_uname(uid, underscores=False):
    # choose base words and strip
    uid = int(uid) + time_salt
    adjective = ADJECTIVES[uid % len(ADJECTIVES)]
    noun = NOUNS[uid % len(NOUNS)]
    # join words as requested
    if underscores:
        uname = adjective + "_" + noun
    else:
        camel_case_adjective = adjective[0].upper() + adjective[1:]
        camel_case_noun = noun[0].upper() + noun[1:]
        uname = (camel_case_adjective + camel_case_noun)
    return uname

def get_ucolour(uid):
    uname = get_uname(uid)
    colour = colours.get(uname)
    if colour is None:
        colour = random.randint(0, max_colour)
        colours[uname] = colour
    return colour