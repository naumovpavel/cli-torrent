package commands

type Command interface {
	Name() string
	Desc() string
	Execute(args ...string) string
}

type command struct {
	name string
	desc string
}

func (c *command) Name() string {
	return c.name
}

func (c *command) Desc() string {
	return c.desc
}
