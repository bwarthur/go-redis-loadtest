package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"

	"github.com/kc0isg/go-redis-loadtest/internal/pkg/docker"
	"github.com/kc0isg/go-redis-loadtest/internal/pkg/rediscmd"
	"github.com/kc0isg/go-redis-loadtest/internal/pkg/utilities"
)

var (
	stopNodesFlag           = flag.Bool("stop-nodes", false, "if true, during execution the master node will be stopped and then later started again")
	metricsAddrFlag         = flag.String("metrics-address", ":8080", "The address to listen on for prometheus scrapping")
	metricsPushAddrFlag     = flag.String("metrics-push-addr", "", "Address to use for prometheus push gateway")
	metricsPushIntervalFlag = flag.Int("metrics-push-interval", 10, "Interval for sending metrics to prometheus push gateway in seconds")
	logFileFlag             = flag.String("log-path", "logrus.log", "File path to use for the log file")

	envoyProxyAddrFlag = flag.String("envoy-addr", "redis-envoy:6379", "Address for envoy proxy")

	stopMasterAfterFlag    = flag.Int("stop-master-after", 60, "Number of seconds to wait before killing master node")
	restartMasterAfterFlag = flag.Int("restart-master-after", 60, "Number of seconds to wait to start master node after stopping it")
	postRestartLengthFlag  = flag.Int("post-restart-length", 30, "Number of seconds to run after restarting master node")

	redisNodes   = &rediscmd.NodeMapFlag{}
	redisNodeMap = map[string]string{"redis-1": "localhost:6381", "redis-2": "localhost:6382", "redis-3": "localhost:6383"}

	redisClient *redis.Client
)

func main() {
	flag.Var(redisNodes, "redis-node", "Redis node name and address, ex. redis-1=localhost:6381")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{})

	file, err := os.OpenFile(*logFileFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	log.Info("Starting Load Test")

	prometheus.Register(rediscmd.Counter)
	prometheus.Register(rediscmd.SuccessCounter)
	prometheus.Register(rediscmd.FailedCounter)
	prometheus.Register(rediscmd.OperationDurationsHistogram)
	prometheus.Register(rediscmd.MTTRHistogram)
	prometheus.Register(prometheus.NewBuildInfoCollector())

	http.Handle("/metricz", promhttp.Handler())
	go func() {
		log.Infof("Starting http listener on %v", *metricsAddrFlag)
		log.Fatal(http.ListenAndServe(*metricsAddrFlag, nil))
	}()

	waitGroup := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	if *metricsPushAddrFlag != "" {
		hostname, err := os.Hostname()
		var jobName string

		if err != nil {
			jobName = "redis_load"
		} else {
			jobName = fmt.Sprintf("redis_load_%s", hostname)
		}

		pusher := push.New(*metricsPushAddrFlag, jobName).Gatherer(prometheus.DefaultGatherer)
		go func() {
			pushOnInterval(ctx, &waitGroup, pusher)
		}()
	}

	redisClient = rediscmd.CreateClient(*envoyProxyAddrFlag, 0)
	rediscmd.CleanKeys(redisClient)

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting Load...")

	go func() {
		rediscmd.RunSetLoop(ctx, &waitGroup, redisClient)
	}()

	go func() {
		rediscmd.RunGetLoop(ctx, &waitGroup, redisClient)
	}()

	go func() {
		rediscmd.RunIncrLoop(ctx, &waitGroup, redisClient)
	}()

	go func() {
		rediscmd.RunLPushLoop(ctx, &waitGroup, redisClient)
	}()

	go func() {
		rediscmd.RunLRangeLoop(ctx, &waitGroup, redisClient)
	}()

	if *stopNodesFlag {
		createFailures(cli)
	} else {
		waitForCtrlC()
	}

	fmt.Println("Stopping Load...")
	cancel()

	waitGroup.Wait()
	fmt.Println("Quiting...")

	processResults()
}

func pushOnInterval(ctx context.Context, wg *sync.WaitGroup, pusher *push.Pusher) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-time.After(time.Duration(*metricsPushIntervalFlag) * time.Second):
			pusher.Push()
			fmt.Println("Pushing Metrics...")

		case <-ctx.Done():
			pusher.Push()
			fmt.Println("pushOnInterval stopping...")
			return
		}
	}
}

func createFailures(cli *client.Client) {

	masterNode, err := rediscmd.FindMasterNode(redisClient, redisNodes)
	if err != nil {
		fmt.Printf("Error getting master nodes: %v\n", err)
	}

	fmt.Printf("Master node: %v\n", masterNode)
	<-time.After(time.Duration(*stopMasterAfterFlag) * time.Second)
	fmt.Println("Stopping master node")
	docker.StopNode(masterNode, cli)

	<-time.After(time.Duration(*restartMasterAfterFlag) * time.Second)
	fmt.Println("Starting master node")
	docker.StartNode(masterNode, cli)

	<-time.After(time.Duration(*postRestartLengthFlag) * time.Second)
}

func processResults() {
	gather, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		fmt.Printf("Error gathering metrics: %v", err)
	}

	opTotal, _ := utilities.GetCounter("redis_operation_total", gather)
	opSuccess, _ := utilities.GetCounter("redis_operation_success_total", gather)
	opFailed, _ := utilities.GetCounter("redis_operation_failed_total", gather)
	opMttr, _ := utilities.GetHistogramSum("redis_operation_mttr_histogram_seconds", gather)

	opAvailability := opSuccess / opTotal

	fmt.Printf("redis_operation_total          => %f\n", opTotal)
	fmt.Printf("redis_operation_success_total  => %f\n", opSuccess)
	fmt.Printf("redis_operation_failed_total   => %f\n", opFailed)
	fmt.Printf("Availability                   => %f\n", opAvailability)
	fmt.Printf("MTTR (seconds)                 => %f\n", opMttr)

	prometheus.WriteToTextfile("stats.txt", prometheus.DefaultGatherer)

	// for _, g := range gather {
	// 	MetricFamilyToText(os.Stdout, g)
	// }

	// for _, g := range gather {
	// 	fmt.Printf("Family Name: %s\n", g.GetName())
	// }

}

func waitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}
