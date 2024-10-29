package core

import (
	"context"
	"errors"
	"github.com/sethvargo/go-retry"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetRetry(ctx context.Context, uri string, retries uint64) ([]byte, error) {
	backoff := retry.WithMaxRetries(retries, retry.NewExponential(3*time.Second))
	return retry.DoValue(ctx, backoff, func(ctx context.Context) ([]byte, error) {
		resp, e := http.Get(uri)
		if e != nil {
			return nil, retry.RetryableError(e)
		}
		if resp.StatusCode != 200 {
			log.Printf("code %v for %v \n", resp.StatusCode, uri)
			return nil, retry.RetryableError(errors.New("non success response " + strconv.FormatInt(int64(resp.StatusCode), 10)))
		}
		body, e := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if body == nil {
			return nil, retry.RetryableError(e)
		}
		return body, nil
	})
}
