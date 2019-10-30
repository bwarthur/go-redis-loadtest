package rediscmd

import (
	"errors"
	"fmt"
	"strings"
)

type NodeMapFlag map[string]string

func (i *NodeMapFlag) String() string {
	var res string
	for k, v := range *i {
		res += fmt.Sprintf("%s=%s", k, v)
	}

	return res
}

func (i *NodeMapFlag) Set(value string) error {
	split := strings.Split(value, "=")
	if len(split) != 2 {
		return errors.New("Cannot parse node")
	}

	(*i)[split[0]] = split[1]

	return nil
}
