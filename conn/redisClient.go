package conn

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func NewRedisClient(host string, port int) (redis.Conn, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {

		return nil, err
	}
	return conn, err
}
