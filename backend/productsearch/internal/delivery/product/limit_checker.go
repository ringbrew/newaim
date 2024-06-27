package product

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"github.com/ringbrew/newaim/productsearch/internal/domain/product"
)

type Limiter struct {
	rds  *redis.Client
	rule map[Aspect]AspectRuleEntry
}

func NewLimiter(ctx *domain.UseCaseContext) *Limiter {
	return &Limiter{
		rds: ctx.Redis,
		rule: map[Aspect]AspectRuleEntry{
			AspectApiKeyAccess: {
				IntervalSec: 10,
				Limit:       10,
			},
			AspectApiKeyInput: {
				IntervalSec: 10,
				Limit:       10,
			},
			AspectApiKeyOutput: {
				IntervalSec: 10,
				Limit:       10,
			},
		},
	}
}

type AspectRuleEntry struct {
	IntervalSec int64
	Limit       int64
}
type Aspect int

const (
	AspectInvalid Aspect = iota
	AspectApiKeyAccess
	AspectApiKeyInput
	AspectApiKeyOutput
)

func (a *Aspect) GenKey(apiKey string, input SearchParam, output []product.Product) (string, error) {
	format := "newaim_product_service_apikey_%s_aspect_%d_%s_limit"

	getDataMd5 := func(input interface{}) (string, error) {
		sData, err := json.Marshal(input)
		if err != nil {
			return "", err
		}

		md5Hash := md5.New()
		_, err = md5Hash.Write(sData)
		if err != nil {
			return "", err
		}
		hashByte := md5Hash.Sum(nil)
		return hex.EncodeToString(hashByte), nil
	}
	switch *a {
	//case
	case AspectApiKeyAccess:
		return fmt.Sprintf(format, apiKey, *a, "access"), nil
	case AspectApiKeyInput:
		dataKey, err := getDataMd5(input)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(format, apiKey, *a, dataKey), nil
	case AspectApiKeyOutput:
		dataKey, err := getDataMd5(output)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(format, apiKey, *a, dataKey), nil
	default:
		return "", nil
	}
}

type CheckLimitInput struct {
	Aspect Aspect
	ApiKey string
	Input  SearchParam
	Output []product.Product
}

func (lc *Limiter) Check(ctx context.Context, input CheckLimitInput) error {
	key, err := input.Aspect.GenKey(input.ApiKey, input.Input, input.Output)
	if err != nil {
		return err
	}

	if key == "" {
		return nil
	}

	c, err := lc.rds.Eval(ctx,
		"local v = redis.call('INCR', KEYS[1]) if v == 1 then redis.call('EXPIRE', KEYS[1], ARGV[1]) end return v",
		[]string{key},
		lc.rule[input.Aspect].IntervalSec).Result()
	if err != nil {
		return err
	}

	count := c.(int64)

	if limit := lc.rule[input.Aspect].Limit; count > limit {
		if count > 20*limit {
			lc.rds.Set(ctx, key, 20*limit, 0)
		}

		return errors.New("fetch forbidden")
	}

	return nil
}
