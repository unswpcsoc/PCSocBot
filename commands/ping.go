package commands

type Ping struct {
	names []string
	desc  string
}

func NewPing() *Ping {
	return *Ping{
		names: []string{"ping", "ping pong"},
		desc:  "Ping!",
	}
}

func (p *Ping) Names() []string {
	return p.names
}

func (p *Ping) Desc() string {
	return p.desc
}

func (p *Ping) Roles() []string {
	return nil
}

func (p *Ping) Channels() []string {
	return nil
}

func (p *Ping) MsgHandle() (Send, error) {
	return nil, nil
}
