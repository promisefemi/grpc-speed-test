package main

import (
	"client/proto"
	"context"
	"flag"
	"fmt"
	pool "github.com/promisefemi/grpc-client-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type counter struct {
	sync.Mutex
	entries                  []int
	currentCount, totalCount int
}

var deviceCount counter

func main() {

	port := flag.String("port", "9000", "Port")
	thread := flag.Int("t", 1, "Thread")
	env := flag.String("e", "", "Env")
	flag.Parse()

	poolConfig := &pool.PoolConfig{
		MaxOpenConnection:     10,
		MaxIdleConnection:     10,
		ConnectionQueueLength: 1000,
		NewClientDuration:     4 * time.Second,
		Address:               fmt.Sprintf(":%s", *port),
		ConfigOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		},
	}
	if *env != "" {
		poolConfig.Address = fmt.Sprintf("cached.tacticallinux.com:%s", *port)
	}
	var deviceId int64 = 6248148189135751453
	poolCon := pool.NewClientPool(poolConfig)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			//fmt.Println("Num of Pool", pool.GetNumberOfOpenConnections())
			deviceCount.Lock()
			fmt.Printf("Get RPC Count: %d -- TimeStamp: %d \n", deviceCount.currentCount, time.Now().Unix())
			fmt.Printf("Total RPC Count: %d \n", deviceCount.totalCount)
			fmt.Printf("Total Number of Connections in Use: %d \n", poolCon.GetNumberOfConnectionsInUse())
			fmt.Printf("Total Number of Available Idle Connections: %d \n\n", poolCon.GetNumberOfIdleConnections())
			deviceCount.entries = append(deviceCount.entries, deviceCount.currentCount)
			deviceCount.currentCount = 0
			deviceCount.Unlock()
		}
	}()

	if *thread <= 1 {
		rpcCall(poolCon, deviceId, &deviceCount, 0)
	} else {
		for i := *thread; i > 0; i-- {
			go rpcCall(poolCon, deviceId, &deviceCount, i)
		}
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		ticker.Stop()
		total := 0
		for _, entry := range deviceCount.entries {
			total += entry
		}
		fmt.Printf("\nAverage request made per second %d\n", total/len(deviceCount.entries))
		os.Exit(0)
	}()
	for {
	}
}

func rpcCall(pool *pool.ClientPool, deviceId int64, deviceCount *counter, i int) {
	for {
		_, err := makeCall(pool, deviceId)
		if err != nil {
			fmt.Println(err)
		} else {
			deviceCount.Lock()
			deviceCount.currentCount++
			deviceCount.totalCount++
			deviceCount.Unlock()
		}
	}
}
func makeCall(pool *pool.ClientPool, deviceId int64) (*proto.Device, error) {
	conn, err := pool.Get()
	if err != nil {
		return nil, fmt.Errorf("error unable to establish connection -- %s\n", err)
	}
	defer conn.Release()
	cachedService := proto.NewCacheClient(conn.Conn)
	deviceInput := &proto.DeviceInput{
		Id: deviceId,
	}
	return cachedService.GetDevice(context.Background(), deviceInput)
}
