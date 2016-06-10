package bot

import "github.com/pajlada/pajbot2/common"

/*
Handle attempts to handle the given message
*/
func (bot *Bot) Handle(msg common.Msg) {
	action := &Action{}
	for _, module := range bot.Modules {
		module.Check(bot, &msg, action)

		if action.Response != "" {
			bot.Say(action.Response)
		}

		if action.Stop {
			return
		}
	}
}