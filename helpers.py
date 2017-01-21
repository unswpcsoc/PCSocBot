class classproperty(object):
    def __init__(self, f):
        self.f = f

    def __get__(self, obj, owner):
        return self.f(owner)

def surround(s, markup):
    return markup + s + markup[::-1]

def bold(s):
    return surround(s, '**')

def underline(s):
    return surround(s, '__')

def code(s):
    return surround(s, '`')

class CommandFailure(Exception):
    pass