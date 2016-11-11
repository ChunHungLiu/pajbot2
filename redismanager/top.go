package redismanager

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/pajlada/pajbot2/common"
)

// Top returns a list of the top users for the given category
func (r *RedisManager) Top(channel string, category string, limit int) []common.User {
	userList := []common.User{}
	const keyF = "%s:users:%s"

	conn := r.Pool.Get()
	defer conn.Close()
	conn.Send("ZREVRANGEBYSCORE",
		fmt.Sprintf(keyF, channel, category),
		"+inf",
		"-inf",
		"LIMIT",
		"0",
		limit)

	conn.Flush()
	res, err := conn.Receive()
	usernames, err := redis.Strings(res, err)
	// this is poorly optimized
	for _, username := range usernames {
		u := common.User{
			Name: username,
		}
		u.LoadDataConn(conn, channel)
		userList = append(userList, u)
	}

	return userList
}
