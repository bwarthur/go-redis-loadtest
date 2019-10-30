package rediscmd

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRunSetLoop(t *testing.T) {
	redisClient := CreateClient("localhost:6379", 0)
	waitGroup := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	go func() { RunSetLoop(ctx, &waitGroup, redisClient) }()

	<-time.After(time.Duration(10) * time.Second)

	cancel()
	waitGroup.Wait()
}

func BenchmarkRunSetLoop(t *testing.T) {

	redisClient := CreateClient("localhost:6379", 0)
	waitGroup := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	go func() { RunSetLoop(ctx, &waitGroup, redisClient) }()

	<-time.After(time.Duration(10) * time.Second)

	cancel()
	waitGroup.Wait()
}
