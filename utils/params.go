package utils

import (
	"strconv"

	"github.com/gobuffalo/buffalo"
)

// GetParam returns string or integer based on the value.
func GetParam(c buffalo.Context, key string) interface{} {
	s := c.Param(key)
	i, err := strconv.Atoi(s)
	if err == nil {
		return i
	}
	return s
}

// GetIntParam extracts and returns an integer value from request parameter.
// if the parameter is not exist or smaller than min or larger than max,
// it returns min value.
func GetIntParam(c buffalo.Context, key string, min, max int) (i int) {
	i, err := strconv.Atoi(c.Param(key))
	if err != nil || i < min {
		return min
	}
	if max > min && i > max {
		return max
	}
	return
}

// GetStringParam extracts and returns a string value from request parameter.
// if the parameter is not exist, it returns failback value.
func GetStringParam(c buffalo.Context, key, failback string) (val string) {
	val = c.Param(key)
	if val == "" {
		return failback
	}
	return
}
