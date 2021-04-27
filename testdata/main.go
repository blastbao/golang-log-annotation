package main

import (
	"context"
	"github.com/blastbao/golang-log-annotation/testdata/log"
)

func main() {
	fn(context.Background(), 1, "2", "3", true)
}

// @Log()
func fn(ctx context.Context, a int, b, c string, d bool) (int, string, string) {
	log.Logger.WithContext(ctx).Infof("#fn executing...")
	return a, b, c
}
