package modules

import (
	"database/sql"
	"strings"

	"github.com/pajlada/pajbot2/bot"
	"github.com/pajlada/pajbot2/command"
	"github.com/pajlada/pajbot2/common"
	"github.com/pajlada/pajbot2/helper"
	"github.com/pajlada/pajbot2/sqlmanager"
)

/*
Command xD
*/
type Command struct {
	commands []command.Command
}

/*
some ideas:
not sure how this would work but id like to keep things simple

$(user1) = usersource displayname
$(sender) = sender of msg
$(1) = raw arg from msg
$(bot) = bot
$(channel) = channel

users have methods, for example $(sender.points):
	.low = lowercase
	.points
	.level
	etc ...

bot has:
	bot.uptime
	bot.version
	etc...

channel has:
	channel.uptime
	channel.title
	channel.game
	channel.subs
	channel.name
	etc...

!points would look like this :
	"$(user1) has $(user1.points) points."

!uptime:
	"$(sender), $(channel.name) has been online for $(channel.uptime) PogChamp"

*/

// Ensure the module implements the interface properly
var _ Module = (*Command)(nil)

func (module *Command) loadCommands(sql *sqlmanager.SQLManager) {
	// Fetch rows from pb_command
	rows, err := sql.Session.Query("SELECT id, channel_id, triggers, response, response_type FROM pb_command")

	if err != nil {
		log.Error("Error fetching commands:", err)
		return
	}

	module.readCommands(rows)
}

// loadCommand loads a command with a given ID
func (module *Command) loadCommand(sql *sqlmanager.SQLManager, commandID int64) {
	// Fetch rows from pb_command
	rows, err := sql.Session.Query("SELECT id, channel_id, triggers, response, response_type FROM pb_command WHERE `id`=?", commandID)

	if err != nil {
		log.Error("Error fetching commands:", err)
		return
	}

	module.readCommands(rows)
}

func (module *Command) readCommands(rows *sql.Rows) {
	for rows.Next() {
		c := command.ReadSQLCommand(rows)
		if c != nil {
			module.AddCommand(c)
		}
	}
}

// Init initializes something
func (module *Command) Init(sql *sqlmanager.SQLManager) {
	module.loadCommands(sql)

	xdCommand := command.TextCommand{
		BaseCommand: command.BaseCommand{
			Triggers: []string{
				"xd",
				"xdlol",
			},
		},
		Response: "pajaSWA",
	}
	module.AddCommand(&xdCommand)
	testCommand := command.NestedCommand{
		BaseCommand: command.BaseCommand{
			Triggers: []string{
				"lul",
				"xdlul",
			},
		},
		Commands: []command.Command{
			&xdCommand,
			&command.TextCommand{
				BaseCommand: command.BaseCommand{
					Triggers: []string{
						"a",
					},
				},
				Response: "pajaSWA a ;P",
			},
			&command.TextCommand{
				BaseCommand: command.BaseCommand{
					Triggers: []string{
						"b",
					},
				},
				Response: "pajaSWA b ;P",
			},
		},
	}
	module.AddCommand(&testCommand)

	// Temporary !admin prefix while it's in development
	adminCommand := &command.NestedCommand{
		BaseCommand: command.BaseCommand{
			Triggers: []string{
				"admin",
			},
		},
		Commands: []command.Command{
			&command.NestedCommand{
				BaseCommand: command.BaseCommand{
					Triggers: []string{
						"add",
					},
				},
				Commands: []command.Command{
					&command.FuncCommand{
						BaseCommand: command.BaseCommand{
							Triggers: []string{
								"command",
							},
							Level: 500,
						},
						Function: module.createCommand,
					},
				},
			},
			&command.NestedCommand{
				BaseCommand: command.BaseCommand{
					Triggers: []string{
						"remove",
					},
				},
				Commands: []command.Command{
					&command.FuncCommand{
						BaseCommand: command.BaseCommand{
							Triggers: []string{
								"command",
							},
							Level: 500,
						},
						Function: module.removeCommand,
					},
				},
			},
			&xdCommand,
			&command.TextCommand{
				BaseCommand: command.BaseCommand{
					Triggers: []string{
						"a",
					},
				},
				Response: "pajaSWA a ;P",
			},
			&command.TextCommand{
				BaseCommand: command.BaseCommand{
					Triggers: []string{
						"b",
					},
				},
				Response: "pajaSWA b ;P",
			},
		},
	}
	module.AddCommand(adminCommand)
}

// AddCommand adds the given command to the list of active commands
func (module *Command) AddCommand(cmd command.Command) {
	module.commands = append(module.commands, cmd)
}

func (module *Command) createCommand(b *bot.Bot, msg *common.Msg, action *bot.Action) {
	// Change to 2 when we remove the !admin prefix
	const triggerLength = 3
	const usageFormat = "Usage: !%s !trigger response"
	triggers := helper.GetTriggersKC(msg.Text)
	if len(triggers) < triggerLength {
		b.Say("Missing arguments")
		return
	}

	if len(triggers) < triggerLength+2 {
		b.Sayf(usageFormat, strings.Join(triggers[:triggerLength], " "))
		return
	}
	// TODO: use an argument parser so we can have --arguments like --silent and --reply --me --cd=0
	arguments := triggers[triggerLength:]

	// TODO: parse multiple triggers (separated by |)
	triggerString := strings.Replace(strings.ToLower(arguments[0]), "!", "", -1)
	if len(triggerString) == 0 {
		b.Sayf(usageFormat, strings.Join(triggers[:triggerLength], " "))
		return
	}
	var triggerList []string
	for _, t := range strings.Split(triggerString, "|") {
		add := true
		if len(t) > 0 {
			for _, eT := range triggerList {
				if t == eT {
					add = false
					break
				}
			}
			if add {
				triggerList = append(triggerList, t)
			}
		}
	}
	b.Sayf("%s", strings.Join(triggerList, ","))
	triggerString = strings.Join(triggerList, "|")

	response := arguments[1:]

	// See if any of the aliases we want to use is already in use
	for _, trigger := range triggerList {
		c := module.getTriggeredCommand("!" + trigger)
		if c != nil {
			b.Sayf("Command !%s is already in use.", trigger)
			return
		}
	}

	sqlCommand := command.SQLCommand{
		ChannelID: 1, // XXX
		Triggers:  triggerString,
		Response:  strings.Join(response, " "),
	}

	b.Sayf("CREATING COMMAND XD: %s - user level: %d", msg.Text, msg.User.Level)
	b.Sayf("Triggers: %s", triggers)
	b.Sayf("Arguments: %s", arguments)
	commandID := sqlCommand.Insert(b.SQL.Session)
	module.loadCommand(b.SQL, commandID)
	b.Say("xD")
}

func (module *Command) removeCommand(b *bot.Bot, msg *common.Msg, action *bot.Action) {
	// Change to 2 when we remove the !admin prefix
	const triggerLength = 3
	const usageFormat = "Usage: !%s !command"
	triggers := helper.GetTriggersKC(msg.Text)
	if len(triggers) < triggerLength {
		b.Say("Missing arguments")
		return
	}

	if len(triggers) < triggerLength+1 {
		b.Sayf(usageFormat, strings.Join(triggers[:triggerLength], " "))
		return
	}
	// TODO: use an argument parser so we can have --arguments like --silent and --reply --me --cd=0
	arguments := triggers[triggerLength:]

	trigger := strings.Replace(strings.ToLower(arguments[0]), "!", "", -1)

	c := module.getTriggeredCommand("!" + trigger)
	if c == nil {
		b.Sayf("No command with trigger !%s", trigger)
		return
	}

	// TODO: Actually remove the command
	b.Sayf("Remove command with trigger !%s", trigger)
	b.Sayf("Command data: %#v", c)
}

func (module *Command) getTriggeredCommand(text string) command.Command {
	m := helper.GetTriggers(text)
	trigger := m[0]

	for _, command := range module.commands {
		if triggered, c := command.IsTriggered(trigger, m, 0); triggered {
			return c
		}
	}
	return nil
}

// Check xD
func (module *Command) Check(b *bot.Bot, msg *common.Msg, action *bot.Action) error {
	if len(msg.Text) == 0 {
		// Do nothing with empty messages
		return nil
	}

	m := helper.GetTriggers(msg.Text)

	if msg.Text[0] != '!' {
		return nil
	}
	c := module.getTriggeredCommand(msg.Text)
	if c != nil {
		// Is the user high level enough to use this command?
		bc := c.GetBaseCommand()
		if bc.Level > msg.User.Level {
			log.Warningf("%s tried to use %s, which requires level %d (he is level %d)",
				msg.User.DisplayName, strings.Join(m, " "), bc.Level, msg.User.Level)
			return nil
		}
		// TODO: Get response first, and skip if the response is nil or something of that sort
		r := c.Run(b, msg, action)
		if r != "" {
			args := strings.Split(msg.Text, " ")
			if len(args) > 1 {
				msg.Args = args[1:]
				log.Debug(msg.Args)
			}
			b.SayFormat(r, msg)
		}
	}
	return nil
}
