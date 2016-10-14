package modules

import (
	"math"
	"strings"

	"github.com/pajlada/pajbot2/bot"
	"github.com/pajlada/pajbot2/common/basemodule"

	"github.com/pajlada/pajbot2/common"
)

// Pyramid module from pajbot 1
type Pyramid struct {
	basemodule.BaseModule

	data      [][]string
	goingDown bool
}

// Ensure the module implements the interface properly
var _ Module = (*Pyramid)(nil)

// NewPyramid xD
func NewPyramid() *Pyramid {
	m := Pyramid{
		BaseModule: basemodule.NewBaseModule(),
	}
	m.ID = "pyramid"
	return &m
}

// Init xD
func (module *Pyramid) Init(bot *bot.Bot) (string, bool) {
	module.SetDefaults("pyramid")
	module.EnabledDefault = true
	module.ParseState(bot.Redis, bot.Channel.Name)

	return "pyramid", isModuleEnabled(bot, "pyramid", true)
}

// DeInit xD
func (module *Pyramid) DeInit(b *bot.Bot) {

}

// Check KKona
func (module *Pyramid) Check(b *bot.Bot, msg *common.Msg, action *bot.Action) error {
	if msg.Type != common.MsgPrivmsg {
		return nil
	}
	msgParts := strings.Split(msg.Text, " ")
	if len(module.data) > 0 {
		curLen := len(msgParts)
		lastLen := len(module.data[len(module.data)-1])
		pyramidThing := module.data[0][0]
		lenDiff := curLen - lastLen

		if math.Abs(float64(lenDiff)) == 1 {
			good := true

			for _, w := range msgParts {
				if w != pyramidThing {
					good = false
					break
				}
			}
			if good {
				module.data = append(module.data, msgParts)
				if lenDiff > 0 {
					if module.goingDown {
						module.data = make([][]string, 0)
						module.goingDown = false
					}
				} else if lenDiff < 0 {
					module.goingDown = true
					if curLen == 1 {
						// a pyramid has been finished
						var peakLen int
						for _, x := range module.data {
							if len(x) > peakLen {
								peakLen = len(x)
							}
						}
						if peakLen > 2 {
							b.SaySafef("%s just finished a %d width %s pyramid PogChamp //",
								msg.User.DisplayName,
								peakLen,
								pyramidThing)
						}
						module.data = make([][]string, 0)
						module.goingDown = false
					}
				}
			} else {
				module.data = make([][]string, 0)
				module.goingDown = false
			}
		} else {
			module.data = make([][]string, 0)
			module.goingDown = false
		}
	}
	if len(msgParts) == 1 && len(module.data) == 0 {
		module.data = append(module.data, msgParts)
	}
	return nil
}
