package redis

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/mediocregopher/radix/v3"
)

// ClientOptions struct contains the options for connecting to redis
type ClientOptions struct {
	Host            string
	Port            string
	Password        string
	MaxRetries      int
	MinRetryBackOff time.Duration
	MaxRetryBackOff time.Duration
	WriteTimeout    time.Duration
	DB              int
}

var dummyHashMap = make(map[string]string)

// Client struct holds connection to redis
type Client struct {
	conn *redis.Client
}

// Clientv2 struct holds pool connection to redis using radix dep
type Clientv2 struct {
	pool *radix.Pool
}

// NewClient method will return a pointer to new client object
func NewClient(opts *ClientOptions) *Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:            opts.Host + ":" + opts.Port,
		Password:        opts.Password,
		DB:              opts.DB,
		MaxRetries:      opts.MaxRetries,
		MinRetryBackoff: opts.MinRetryBackOff,
		MaxRetryBackoff: opts.MaxRetryBackOff,
		WriteTimeout:    opts.WriteTimeout,
	})
	var client = &Client{conn: redisClient}
	return client
}

// NewV2Client will return the pool connection to radix object
func NewV2Client(opts *ClientOptions) *Clientv2 {
	// Ref: https://github.com/mediocregopher/radix/blob/master/radix.go#L107
	customConnFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialTimeout(500*time.Millisecond),
			radix.DialAuthPass(opts.Password),
			radix.DialSelectDB(opts.DB),
		)
	}

	rclient, _ := radix.NewPool("tcp", opts.Host+":"+opts.Port, 15, radix.PoolConnFunc(customConnFunc))
	var client = &Clientv2{pool: rclient}
	return client
}

// GetConn returns a pointer to the underlying redis library
func (c *Client) GetConn() *redis.Client {
	return c.conn
}

// HIncrBy will increment a hash map key
func (c *Client) HIncrBy(key, field string, inc int64) int64 {
	resp := c.conn.HIncrBy(key, field, inc)
	result, _ := resp.Result()
	return result
}

// HGetAll will return the hash map
func (c *Client) HGetAll(key string) map[string]string {
	resp := c.conn.HGetAll(key)
	result, err := resp.Result()
	if err != nil {
		return dummyHashMap
	}
	return result
}

// Del method will remove single key from redis
func (c *Client) Del(key string) {
	c.conn.Del(key)
}

// DelMulti method will remove multiple keys from redis
func (c *Client) DelMulti(keys []string) {
	c.conn.Del(keys...)
}

// HIncrBy will increment a hash map key
func (c *Clientv2) HIncrBy(key, field string, inc int64) {
	val := strconv.Itoa(int(inc))
	c.pool.Do(radix.Cmd(nil, "HINCRBY", key, field, val))
}

// HIncrByFloat will increment a hash map key
func (c *Clientv2) HIncrByFloat(key, field string, inc float64) {
	val := fmt.Sprintf("%f", inc)
	c.pool.Do(radix.Cmd(nil, "HINCRBYFLOAT", key, field, val))
}

// SCard will get the size of set
func (c *Clientv2) SCard(key string) int {
	var count int
	c.pool.Do(radix.Cmd(&count, "SCARD", key))
	return count
}

// SIsMember will will check if value is in the set
func (c *Clientv2) SIsMember(key, val string) int {
	var isMember int
	c.pool.Do(radix.Cmd(&isMember, "SISMEMBER", key, val))
	return isMember
}

// SAdd will add the member to the set
func (c *Clientv2) SAdd(key, field string) int {
	var success int
	c.pool.Do(radix.Cmd(&success, "SADD", key, field))
	return success
}

// Close method closes the redis connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Close method closes the redis connection
func (c *Clientv2) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}
