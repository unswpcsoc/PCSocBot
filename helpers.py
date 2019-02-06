from datetime import datetime, timedelta

class classproperty(object):
    def __init__(self, f):
        self.f = f

    def __get__(self, obj, owner):
        return self.f(owner)

def surround(s, markup):
    return markup + s + markup[::-1]

def italics(s):
    return surround(s, '*')

def bold(s):
    return surround(s, '**')

def underline(s):
    return surround(s, '__')

def code(s):
    return surround(s, '`')

def spoiler(s):
    return surround(s, '||')

def codeblock(s):
    return '```\n' + s + '\n```'

def at(s):
    return "<@%s>" % s

def noembed(s):
    return "<%s>" % s

def chan(s):
    return "<#%s>" % s

def nick(u):
    try:
        return u.nick or str(u).split("#")[0]
    except AttributeError:
        return str(u).split("#")[0]

def timestamp():
    return datetime.now().strftime("[%Y-%m-%d %H:%M:%S]")

def duration(s):
    return str(timedelta(seconds=int(s)))

def is_good_response(resp):
    """ For music.py autosuggest fetching
    Returns True if the response seems to be HTML, False otherwise.
    """
    content_type = resp.headers['Content-Type'].lower()
    return (resp.status_code == 200 
            and content_type is not None 
            and content_type.find('html') > -1)

class CommandFailure(Exception):
    pass

class BadHTMLError(Exception):
    def __init__(self, message):
        self.message = message
