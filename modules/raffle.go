package modules

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pajlada/pajbot2/bot"
	"github.com/pajlada/pajbot2/common"
	"github.com/pajlada/pajbot2/common/basemodule"
)

// Raffle module
type Raffle struct {
	basemodule.BaseModule
	bot    *bot.Bot
	users  []string
	length time.Duration
	points int

	MinPoints *int64 `json:"min_points"`
	MaxPoints *int64 `json:"max_points"`
}

// Ensure the module implements the interface properly
var _ Module = (*Raffle)(nil)

// NewRaffle xD
func NewRaffle() *Raffle {
	m := Raffle{
		BaseModule: basemodule.NewBaseModule(),
	}
	m.ID = "raffle"
	m.EnabledDefault = true
	return &m
}

// Init xD
func (module *Raffle) Init(bot *bot.Bot) (string, bool) {
	module.ParseState(bot.Redis, bot.Channel.Name)

	module.bot = bot
	module.parseSettings(module.FetchSettings(bot.Redis, bot.Channel.Name))

	return "raffle", isModuleEnabled(bot, "raffle", true)
}

// DeInit xD
func (module *Raffle) DeInit(bot *bot.Bot) {
}

// Check xD
func (module *Raffle) Check(b *bot.Bot, msg *common.Msg, action *bot.Action) error {
	if !strings.HasPrefix(msg.Text, "!") {
		return nil
	}
	spl := strings.Split(msg.Text, " ")
	trigger := strings.ToLower(spl[0])
	switch trigger {
	case "!raffel":
		if msg.User.Level < 500 {
			return nil
		}
		if module.points != 0 {
			b.Sayf("%s there is already a raffle running OMGScoots", msg.User.DisplayName)
			return nil
		}
		if len(spl) < 2 {
			module.newRaffle([]string{})
			return nil
		}
		module.newRaffle(spl[1:])
	case "!join":
		var joined bool
		for _, user := range module.users {
			if user == msg.User.Name {
				joined = true
			}
		}
		if !joined {
			module.users = append(module.users, msg.User.Name)
		}
	}
	return nil
}

func (module *Raffle) newRaffle(args []string) {
	switch len(args) {
	case 0:
		go module.startRaffle(60*time.Second, 1000)
	case 1:
		points, err := strconv.Atoi(args[0])
		if err != nil {
			go module.startRaffle(60*time.Second, 1000)
			return
		}
		go module.startRaffle(60*time.Second, points)
	default:
		points, err := strconv.Atoi(args[0])
		if err != nil {
			go module.startRaffle(60*time.Second, 1000)
			return
		}
		length, err := strconv.Atoi(args[1])
		if err != nil {
			go module.startRaffle(60*time.Second, points)
			return
		}
		go module.startRaffle(time.Duration(length)*time.Second, points)
	}
}

func (module *Raffle) startRaffle(length time.Duration, points int) {
	module.points = points
	module.length = length
	module.bot.Sayf("a raffle has begun for %d points pajaDank ends in %.f seconds KKaper",
		points, length.Seconds())
	step := length / 4
	steps := []float64{0.75, 0.5, 0.25}
	for _, i := range steps {
		time.Sleep(step)
		module.bot.Sayf("the raffle for %d points ends in %.f seconds, enter by typing !join OpieOP",
			points, length.Seconds()*i)
	}
	time.Sleep(step)
	if len(module.users) < 1 {
		module.bot.Say("no one entered the raffle LUL")
		module.reset()
		return
	}
	winners, pts := module.getWinners()
	var winnersString string
	for _, w := range winners {
		winnersString += w + ","
		module.bot.Redis.IncrPoints(module.bot.Channel.Name, w, pts)
	}
	module.bot.Sayf("%s won %d each PogChamp", winnersString, pts)
	module.reset()
}

func (module *Raffle) reset() {
	module.points = 0
	module.users = []string{}
	module.length = time.Second * 0
}

func (module *Raffle) getWinners() ([]string, int) {
	var winners []string
	winnerCount := 1
	if len(module.users) > 5 {
		winnerCount = len(module.users) / 5
	}
	for i := 0; i < winnerCount; i++ {
		r := rand.Intn(len(module.users))
		u := module.users[r]
		var isWinner bool
		for _, user := range winners {
			if u == user {
				isWinner = true
			}
		}
		if !isWinner {
			winners = append(winners, u)
		}
	}
	return winners, module.points / len(winners)
}

func (module *Raffle) getMinPoints() int {
	if module.MinPoints == nil {
		return 50
	}

	return int(*module.MinPoints)
}

func (module *Raffle) getMaxPoints() int {
	if module.MaxPoints == nil {
		return 50
	}

	return int(*module.MaxPoints)
}

func (module *Raffle) parseSettings(jsonData []byte) {
	json.Unmarshal(jsonData, module)
}
