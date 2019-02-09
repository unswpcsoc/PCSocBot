from commands.base import Command
from helpers import *

from requests import get
from requests.exceptions import RequestException
from contextlib import closing
from bs4 import BeautifulSoup

SUGG_CLASS = " content-link spf-link yt-uix-sessionlink spf-link "
YT_PREFIX = 'https://www.youtube.com'


class Autosuggest(Command):
    desc = "Takes a youtube link and returns the autosuggested links."

    def eval(self, url, count=1):
        content = None
        if not url.startswith(YT_PREFIX):
            raise CommandFailure("Please enter a valid YouTube url!")

        try:
            count = int(count)
        except TypeError:
            raise CommandFailure("Please use an integer for the count!")

        # Get html response from url
        try:
            with closing(get(url, stream=True)) as resp:
                if is_good_response(resp):
                    content = resp.content
                else:
                    raise CommandFailure("Bad Response from %s" % url)

        except RequestException as e:
            raise CommandFailure(
                "Error during requests to %s : %s" % (url, str(e)))

        # (try) Make soup
        try:
            html = BeautifulSoup(content, 'html.parser')
        except BadHTMLError as e:
            raise CommandFailure(e.message)

        # Find autosuggest results
        #results = []
        # for a in html.find_all('a', class_=SUGG_CLASS, limit=count):
        #entry = {'url':YT_PREFIX + a['href'], 'title':a['title']}
        # results.append(entry)

        # Return the results
        #out = bold("Autosuggested results:\n")
        # for result in results:
            #old = out
            #out += result['title'] + "\nLink: " + noembed(result['url']) + "\n"
            # Check for 2000 char limit
            # if len(out) >= 2000:
            #self.send_message(self.message.channel, old)
            #out = result['title'] + "\nLink: " + noembed(result['url']) + "\n"

        result = html.find('a', class_=SUGG_CLASS)
        out = "First Result:\n" + result['title'] + "\nLink: " + \
            noembed(YT_PREFIX + result['href']) + "\n"
        return out
