"""Username generator.

A command to generate a username from an int uid.
Modified from https://pypi.org/project/username-generator/

"""

from .wordlists import ADJECTIVES, NOUNS
import time

timeSalt = int(time.time())
print("username_generator: timeSalt is " + str(timeSalt))

def get_uname(uid, underscores=False):
    # choose base words and strip
    uid = int(uid) + timeSalt
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