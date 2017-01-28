from discord import Embed


class EmbedTable(Embed):
    def __init__(self, fields, table, user=None, *args, **kwargs):
        super().__init__(*args, **kwargs)

        table = zip(*table)
        for field, column in zip(fields, table):
            self.add_field(name=field, value="\n".join(column), inline=True)

        if user is not None:
            self.set_author(name=user.name, icon_url=user.avatar_url)