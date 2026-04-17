package actions

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/itzfelixv/evmgo/internal/rpc"
)

func ParseRawParams(args []string) ([]any, error) {
	params := make([]any, 0, len(args))
	for _, arg := range args {
		var value any
		dec := json.NewDecoder(strings.NewReader(arg))
		dec.UseNumber()
		if err := dec.Decode(&value); err == nil {
			if _, err := dec.Token(); err == io.EOF {
				params = append(params, value)
				continue
			}
		}
		params = append(params, arg)
	}
	return params, nil
}

func CallRaw(ctx context.Context, client *rpc.Client, method string, rawArgs []string) (any, error) {
	params, err := ParseRawParams(rawArgs)
	if err != nil {
		return nil, err
	}

	var result any
	if err := client.Call(ctx, method, params, &result); err != nil {
		return nil, err
	}

	return result, nil
}
