package realtime

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	addr     string
	password string
	db       int
	timeout  time.Duration
}

func NewClient(addr, password string, db int) *Client {
	if strings.TrimSpace(addr) == "" {
		addr = "127.0.0.1:6379"
	}

	return &Client{
		addr:     addr,
		password: password,
		db:       db,
		timeout:  3 * time.Second,
	}
}

func (c *Client) Ping() error {
	reply, err := c.do("PING")
	if err != nil {
		return err
	}

	if text, ok := reply.(string); ok && strings.EqualFold(text, "PONG") {
		return nil
	}

	return fmt.Errorf("unexpected ping response: %v", reply)
}

func (c *Client) Get(key string) (string, bool, error) {
	reply, err := c.do("GET", key)
	if err != nil {
		return "", false, err
	}
	if reply == nil {
		return "", false, nil
	}
	value, ok := reply.(string)
	if !ok {
		return "", false, fmt.Errorf("unexpected GET response type %T", reply)
	}
	return value, true, nil
}

func (c *Client) SetEX(key string, seconds int, value string) error {
	_, err := c.do("SETEX", key, strconv.Itoa(seconds), value)
	return err
}

func (c *Client) SetNXEX(key string, seconds int, value string) (bool, error) {
	reply, err := c.do("SET", key, value, "EX", strconv.Itoa(seconds), "NX")
	if err != nil {
		return false, err
	}
	if reply == nil {
		return false, nil
	}
	text, ok := reply.(string)
	if !ok {
		return false, fmt.Errorf("unexpected SET NX response type %T", reply)
	}
	return strings.EqualFold(text, "OK"), nil
}

func (c *Client) Exists(key string) (bool, error) {
	reply, err := c.do("EXISTS", key)
	if err != nil {
		return false, err
	}
	count, err := toInt64(reply)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *Client) Incr(key string) (int64, error) {
	reply, err := c.do("INCR", key)
	if err != nil {
		return 0, err
	}
	return toInt64(reply)
}

func (c *Client) HSet(key string, fields map[string]string) error {
	// Redis 3.x does not support multi-field HSET.
	// Use HMSET for backward compatibility with local Windows Redis builds.
	args := []string{"HMSET", key}
	for field, value := range fields {
		args = append(args, field, value)
	}
	_, err := c.do(args...)
	return err
}

func (c *Client) HGetAll(key string) (map[string]string, error) {
	reply, err := c.do("HGETALL", key)
	if err != nil {
		return nil, err
	}
	items, ok := reply.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected HGETALL response type %T", reply)
	}
	result := make(map[string]string, len(items)/2)
	for i := 0; i+1 < len(items); i += 2 {
		field, ok := items[i].(string)
		if !ok {
			return nil, errors.New("invalid hash field type")
		}
		value, ok := items[i+1].(string)
		if !ok {
			return nil, errors.New("invalid hash value type")
		}
		result[field] = value
	}
	return result, nil
}

func (c *Client) Del(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	args := append([]string{"DEL"}, keys...)
	_, err := c.do(args...)
	return err
}

func (c *Client) SAdd(key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	args := append([]string{"SADD", key}, members...)
	_, err := c.do(args...)
	return err
}

func (c *Client) SCard(key string) (int64, error) {
	reply, err := c.do("SCARD", key)
	if err != nil {
		return 0, err
	}
	return toInt64(reply)
}

func (c *Client) ZAdd(key string, score int64, member string) error {
	_, err := c.do("ZADD", key, strconv.FormatInt(score, 10), member)
	return err
}

func (c *Client) ZRevRank(key, member string) (int64, bool, error) {
	reply, err := c.do("ZREVRANK", key, member)
	if err != nil {
		return 0, false, err
	}
	if reply == nil {
		return 0, false, nil
	}
	rank, err := toInt64(reply)
	return rank, err == nil, err
}

func (c *Client) ZScore(key, member string) (int64, bool, error) {
	reply, err := c.do("ZSCORE", key, member)
	if err != nil {
		return 0, false, err
	}
	if reply == nil {
		return 0, false, nil
	}
	text, ok := reply.(string)
	if !ok {
		return 0, false, fmt.Errorf("unexpected ZSCORE response type %T", reply)
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, false, err
	}
	return int64(value), true, nil
}

func (c *Client) ZRevRangeWithScores(key string, start, stop int) ([]ScoredMember, error) {
	reply, err := c.do("ZREVRANGE", key, strconv.Itoa(start), strconv.Itoa(stop), "WITHSCORES")
	if err != nil {
		return nil, err
	}
	items, ok := reply.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected ZREVRANGE response type %T", reply)
	}
	result := make([]ScoredMember, 0, len(items)/2)
	for i := 0; i+1 < len(items); i += 2 {
		member, ok := items[i].(string)
		if !ok {
			return nil, errors.New("invalid zset member type")
		}
		scoreText, ok := items[i+1].(string)
		if !ok {
			return nil, errors.New("invalid zset score type")
		}
		scoreValue, err := strconv.ParseFloat(scoreText, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, ScoredMember{
			Member: member,
			Score:  int64(scoreValue),
		})
	}
	return result, nil
}

func (c *Client) RPush(key string, values ...string) error {
	if len(values) == 0 {
		return nil
	}
	args := append([]string{"RPUSH", key}, values...)
	_, err := c.do(args...)
	return err
}

func (c *Client) LTrim(key string, start, stop int) error {
	_, err := c.do("LTRIM", key, strconv.Itoa(start), strconv.Itoa(stop))
	return err
}

func (c *Client) LRange(key string, start, stop int) ([]string, error) {
	reply, err := c.do("LRANGE", key, strconv.Itoa(start), strconv.Itoa(stop))
	if err != nil {
		return nil, err
	}
	items, ok := reply.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected LRANGE response type %T", reply)
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, errors.New("invalid list item type")
		}
		result = append(result, text)
	}
	return result, nil
}

func (c *Client) Eval(script string, keys []string, args []string) (string, error) {
	command := []string{"EVAL", script, strconv.Itoa(len(keys))}
	command = append(command, keys...)
	command = append(command, args...)

	reply, err := c.do(command...)
	if err != nil {
		return "", err
	}
	text, ok := reply.(string)
	if !ok {
		return "", fmt.Errorf("unexpected EVAL response type %T", reply)
	}
	return text, nil
}

type ScoredMember struct {
	Member string
	Score  int64
}

func (c *Client) do(args ...string) (any, error) {
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	if c.password != "" {
		if err := writeCommand(conn, "AUTH", c.password); err != nil {
			return nil, err
		}
		if _, err := readRESP(reader); err != nil {
			return nil, err
		}
	}
	if c.db > 0 {
		if err := writeCommand(conn, "SELECT", strconv.Itoa(c.db)); err != nil {
			return nil, err
		}
		if _, err := readRESP(reader); err != nil {
			return nil, err
		}
	}

	if err := writeCommand(conn, args...); err != nil {
		return nil, err
	}

	return readRESP(reader)
}

func writeCommand(conn net.Conn, args ...string) error {
	if _, err := fmt.Fprintf(conn, "*%d\r\n", len(args)); err != nil {
		return err
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(arg), arg); err != nil {
			return err
		}
	}
	return nil
}

func readRESP(reader *bufio.Reader) (any, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	line, err := readLine(reader)
	if err != nil {
		return nil, err
	}

	switch prefix {
	case '+':
		return line, nil
	case '-':
		return nil, errors.New(line)
	case ':':
		return strconv.ParseInt(line, 10, 64)
	case '$':
		size, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		if size == -1 {
			return nil, nil
		}
		data := make([]byte, size+2)
		if _, err := io.ReadFull(reader, data); err != nil {
			return nil, err
		}
		return string(data[:size]), nil
	case '*':
		count, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		if count == -1 {
			return nil, nil
		}
		result := make([]any, 0, count)
		for i := 0; i < count; i++ {
			value, err := readRESP(reader)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported RESP prefix %q", prefix)
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unexpected integer type %T", value)
	}
}
