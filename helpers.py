from datetime import datetime

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

def codeblock(s):
    return '```\n' + s + '\n```'

def at(s):
    return "<@%s>" % s

def noembed(s):
    return "<%s>" % s

def chan(s):
    return "<#%s>" % s

def nick(m):
    return m.nick or str(m).split("#")[0]

def timestamp():
    return datetime.now().strftime("[%Y-%m-%d %H:%M:%S]")

def duration(player):
    return str(datetime.timedelta(seconds=int(player.duration)))

def is_good_response(resp):
    """ For music.py autosuggest fetching
    Returns True if the response seems to be HTML, False otherwise.
    """
    content_type = resp.headers['Content-Type'].lower()
    return (resp.status_code == 200 
            and content_type is not None 
            and content_type.find('html') > -1)

def log_error(e):
    """ Generic Error Logger
    Print to a logfile
    """
    with open("log", "a") as f:
        f.write("log", "[" + datetime.now().ctime() + "] " + e)
    finally:
        f.close()

class CommandFailure(Exception):
    pass
