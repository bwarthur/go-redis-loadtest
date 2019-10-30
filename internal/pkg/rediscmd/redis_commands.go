package rediscmd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	Counter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "redis_operation_total",
		Help: "Total number of commands sent to redis",
	}, []string{"operation"})
	SuccessCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "redis_operation_success_total",
		Help: "Total number of successful commands sent to redis",
	}, []string{"operation"})
	FailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "redis_operation_failed_total",
		Help: "Total number of failed commands sent to redis",
	}, []string{"operation"})
	OperationDurationsHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "redis_operation_durations_histogram_seconds",
		Help:    "Duration of commands sent to redis in seconds",
		Buckets: []float64{0.000001, 0.000005, 0.00001, 0.00005, 0.0001, 0.0005, 0.001, 0.005, 1},
	}, []string{"operation"})
	MTTRHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "redis_operation_mttr_histogram_seconds",
		Help:    "MTTR in seconds for sending commands to redis",
		Buckets: []float64{0.000001, 0.00001, 0.0001, 0.001, 0.01, 0.1, 1},
	}, []string{"operation"})
)

// CreateClient Creates a new Redis Client based on provided address and default db number
func CreateClient(addr string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	return client
}

// CleanKeys Deletes keys being used for each of the tests
func CleanKeys(redisClient *redis.Client) {
	redisClient.Del("SetKey", "GetKey", "LPushKey", "IncrKey")
}

// FindMasterNode Checks each of the provided nodes to see which one is set as master
func FindMasterNode(redisClient *redis.Client, redisNodeMap *NodeMapFlag) (string, error) {
	for k, v := range *redisNodeMap {
		res, err := CreateClient(v, 0).Do("ROLE").Result()
		if err != nil {
			log.WithError(err).Errorf("Error checking role for '%v'", v)
			return "", err
		}

		role := res.([]interface{})[0].(string)

		if fmt.Sprintf("%v", role) == "master" {
			return k, nil
		}
	}

	log.Error("Unable to find master node")
	return "", errors.New("Unable to find master node")
}

// RunSetLoop Runs redis SET command in a loop until context done is triggered
func RunSetLoop(ctx context.Context, wg *sync.WaitGroup, redisClient *redis.Client) {
	wg.Add(1)
	defer wg.Done()

	iter := 0
	errFlag := false
	var errTime time.Time

	for {
		select {
		default:
			iter++
			start := time.Now()

			err := redisClient.Set("SetKey", "value", 0).Err()

			duration := time.Since(start)
			OperationDurationsHistogram.WithLabelValues("set").Observe(duration.Seconds())

			Counter.WithLabelValues("set").Inc()

			if err != nil {
				log.WithField("iter", iter).WithError(err).Error("Error running SET command")

				errFlag = true
				errTime = time.Now()

				FailedCounter.WithLabelValues("set").Inc()
				//fmt.Println(err)
			} else {
				SuccessCounter.WithLabelValues("set").Inc()

				if errFlag {
					errDuration := time.Since(errTime)
					log.WithField("iter", iter).Infof("SET MTTR: %v", errDuration.Seconds())
					MTTRHistogram.WithLabelValues("set").Observe(errDuration.Seconds())
					errFlag = false
				}
			}

		case <-ctx.Done():
			fmt.Println("runSetLoop stopping...")
			return
		}
	}
}

// RunGetLoop Runs redis GET command in a loop until context done is triggered
func RunGetLoop(ctx context.Context, wg *sync.WaitGroup, redisClient *redis.Client) {
	wg.Add(1)
	defer wg.Done()

	iter := 0
	errFlag := false
	var errTime time.Time

	setKeyErr := redisClient.Set("GetKey", "GetValue", 0).Err()
	if setKeyErr != nil {
		fmt.Println(setKeyErr)
	}

	for {
		select {
		default:
			iter++
			start := time.Now()

			err := redisClient.Get("GetKey").Err()

			duration := time.Since(start)
			OperationDurationsHistogram.WithLabelValues("get").Observe(duration.Seconds())

			Counter.WithLabelValues("get").Inc()

			if err != nil {
				log.WithField("iter", iter).WithError(err).Error("Error running GET command")

				errFlag = true
				errTime = time.Now()

				FailedCounter.WithLabelValues("get").Inc()
				//fmt.Println(err)
			} else {
				SuccessCounter.WithLabelValues("get").Inc()

				if errFlag {
					errDuration := time.Since(errTime)
					log.WithField("iter", iter).Infof("GET MTTR: %v", errDuration.Seconds())
					MTTRHistogram.WithLabelValues("get").Observe(errDuration.Seconds())
					errFlag = false
				}
			}

		case <-ctx.Done():
			fmt.Println("runGetLoop stopping...")
			return
		}
	}
}

// RunIncrLoop Runs redis INCR command in a loop until context done is triggered
func RunIncrLoop(ctx context.Context, wg *sync.WaitGroup, redisClient *redis.Client) {
	wg.Add(1)
	defer wg.Done()

	iter := 0
	errFlag := false
	var errTime time.Time

	for {
		select {
		default:
			iter++
			start := time.Now()

			err := redisClient.Incr("IncrKey").Err()

			duration := time.Since(start)
			OperationDurationsHistogram.WithLabelValues("incr").Observe(duration.Seconds())

			Counter.WithLabelValues("incr").Inc()

			if err != nil {
				log.WithField("iter", iter).WithError(err).Error("Error running INCR command")

				errFlag = true
				errTime = time.Now()

				FailedCounter.WithLabelValues("incr").Inc()
				//fmt.Println(err)
			} else {
				SuccessCounter.WithLabelValues("incr").Inc()

				if errFlag {
					errDuration := time.Since(errTime)
					log.WithField("iter", iter).Infof("INCR MTTR: %v", errDuration.Seconds())
					MTTRHistogram.WithLabelValues("incr").Observe(errDuration.Seconds())
					errFlag = false
				}
			}

		case <-ctx.Done():
			fmt.Println("runIncrLoop stopping...")
			return
		}
	}
}

// RunLPushLoop Runs redis LPUSH command in a loop until context done is triggered
func RunLPushLoop(ctx context.Context, wg *sync.WaitGroup, redisClient *redis.Client) {
	wg.Add(1)
	defer wg.Done()

	iter := 0
	errFlag := false
	var errTime time.Time

	for {
		select {
		default:
			iter++
			start := time.Now()

			err := redisClient.LPush("LPushKey", "Value").Err()

			duration := time.Since(start)
			OperationDurationsHistogram.WithLabelValues("lpush").Observe(duration.Seconds())

			Counter.WithLabelValues("lpush").Inc()

			if err != nil {
				log.WithField("iter", iter).WithError(err).Error("Error running LPUSH command")

				errFlag = true
				errTime = time.Now()

				FailedCounter.WithLabelValues("lpush").Inc()
				//fmt.Println(err)
			} else {
				SuccessCounter.WithLabelValues("lpush").Inc()

				if errFlag {
					errDuration := time.Since(errTime)
					log.WithField("iter", iter).Infof("LPUSH MTTR: %v", errDuration.Seconds())
					MTTRHistogram.WithLabelValues("lpush").Observe(errDuration.Seconds())
					errFlag = false
				}
			}

		case <-ctx.Done():
			fmt.Println("runLPushLoop stopping...")
			return
		}
	}
}

// RunLRangeLoop Runs redis LRANGE command in a loop until context done is triggered
func RunLRangeLoop(ctx context.Context, wg *sync.WaitGroup, redisClient *redis.Client) {
	wg.Add(1)
	defer wg.Done()

	iter := 0
	errFlag := false
	var errTime time.Time

	for i := 0; i < 1000; i++ {
		redisClient.LPush("LRangeKey", i)
	}

	for {
		select {
		default:
			iter++
			start := time.Now()

			err := redisClient.LRange("LRangeKey", 0, -1).Err()

			duration := time.Since(start)
			OperationDurationsHistogram.WithLabelValues("lrange").Observe(duration.Seconds())

			Counter.WithLabelValues("lrange").Inc()

			if err != nil {
				log.WithField("iter", iter).WithError(err).Error("Error running LRANGE command")

				errFlag = true
				errTime = time.Now()

				FailedCounter.WithLabelValues("lrange").Inc()
				//fmt.Println(err)
			} else {
				SuccessCounter.WithLabelValues("lrange").Inc()

				if errFlag {
					errDuration := time.Since(errTime)
					log.WithField("iter", iter).Infof("LRANGE MTTR: %v", errDuration.Seconds())
					MTTRHistogram.WithLabelValues("lrange").Observe(errDuration.Seconds())
					errFlag = false
				}
			}

		case <-ctx.Done():
			fmt.Println("runLRangeLoop stopping...")
			return
		}
	}
}
