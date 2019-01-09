from commands.base import Command
from discord import Embed
from helpers import CommandFailure

import bs4, re, requests

def subject_details(code):
    page = requests.get('https://www.handbook.unsw.edu.au/undergraduate/courses/2019/' + code)

    if page.status_code != 200:
        page = requests.get('https://www.handbook.unsw.edu.au/postgraduate/courses/2019/' + code)

        if page.status_code != 200:
            return None

    soup = bs4.BeautifulSoup(page.text, features='lxml')

    course_title = soup.find('span', attrs={'data-hbui' : 'module-title'}).string.strip()
    course_offerings = soup.find('strong', string=re.compile('Offering Terms')).parent.contents[3].string.strip()

    course_conditions_tag = soup.find('div', id='readMoreSubjectConditions')

    course_conditions = 'None'

    if course_conditions_tag is not None:
        for s in course_conditions_tag.contents[1].contents[1].strings:
            course_conditions += s

    course_conditions = course_conditions.strip()

    course_description_tag = soup.find('div', id='readMoreIntro')

    course_description = ''

    if len(course_description_tag.contents[1].contents) > 1 and course_description_tag.contents[1].contents[1].name == 'p':
        for s in course_description_tag.contents[1].contents[1].strings:
            course_description += s
    else:
        course_description = course_description_tag.contents[1].contents[0].string

    course_description = course_description.strip()

    return {
        'title' : course_title,
        'description' : course_description,
        'offerings' : course_offerings,
        'conditions' : course_conditions,
        'link' : 'https://www.handbook.unsw.edu.au/undergraduate/courses/2019/' + code
    }

class Handbook(Command):
    desc = "This command scrapes entries in the UNSW handbook"

    def eval(self, arg):

        if not re.search(r'^[a-zA-Z]{4}[0-9]{4}$', arg):
            raise CommandFailure('Incorrectly formatted course code: ' + arg)

        course = subject_details(arg)
        if course is not None:
            return 'Course ' + arg + ' could not be found'

        ret = Embed(title=course['title'], description=course['description'], url=course['link'], color=self.EMBED_COLOR)
        ret.add_field(name='Offering Terms', value=course['offerings'])
        ret.add_field(name='Enrolment Conditions', value=course['conditions'])

        return ret