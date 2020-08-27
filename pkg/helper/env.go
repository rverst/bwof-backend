package helper

import (
	"os"
	"strconv"
)

// GetBoolEnv retrieves the value of the environment variable named by the key
// as bool. It returns the `def` value if the environment variable is empty, unset or
// can't be parsed as bool.
func GetBoolEnv(key string, def bool) bool {
	e := os.Getenv(key)
	if e == "" {
		return def
	}
	r, err := strconv.ParseBool(e)
	if err != nil {
		return def
	}
	return r
}

// GetStringEnv retrieves the value of the environment variable named by the key
// as string. It returns the `def` value if the environment variable is empty, or
// unset.
func GetStringEnv(key string, def string) string {
	e := os.Getenv(key)
	if e == "" {
		return def
	}
	return e
}

// GetInt64Env retrieves the value of the environment variable named by the key
// as int64. It returns the `def` value if the environment variable is empty, unset or
// can't be parsed as int64.
func GetInt64Env(key string, def int64) int64 {
	e := os.Getenv(key)
	if e == "" {
		return def
	}
	i, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return def
	}
	return i
}

// GetUint16Env retrieves the value of the environment variable named by the key
// as uint16. It returns the `def` value if the environment variable is empty, unset or
// can't be parsed as uint16.
func GetUint16Env(key string, def uint16) uint16 {
	e := os.Getenv(key)
	if e == "" {
		return def
	}
	i, err := strconv.ParseUint(e, 10, 16)
	if err != nil {
		return def
	}
	return uint16(i)
}
