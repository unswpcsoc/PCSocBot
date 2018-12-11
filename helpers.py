from datetime import datetime

class classproperty(object):
    def __init__(self, f):
        self.f = f

    def __get__(self, obj, owner):
        return self.f(owner)

def surround(s, markup):
    return f'{markup}{s}{markup[::-1]}'

def bold(s):
    return surround(s, '**')

def underline(s):
    return surround(s, '__')

def code(s):
    return surround(s, '`')

def codeblock(s):
    return surround(s, '```\n')

def at(s):
    return f'<@{s}>'

def noembed(s):
    return f'<{s}>'

def chan(s):
    return f'<#{s}>'

def nick(m):
    if m.nick is None:
        # Default username
        return str(m).split("#")[0]
    return m.nick

def timestamp():
    return datetime.now().strftime("[%Y-%m-%d %H:%M:%S]")

class CommandFailure(Exception):
    pass
