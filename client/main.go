package main

import (
	"client/proto"
	"context"
	"flag"
	"fmt"
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
	connToOpen := flag.Int("c", 1, "Thread")
	env := flag.String("e", "", "Env")
	flag.Parse()

	address := fmt.Sprintf(":%s", *port)
	if *env != "" {
		address = fmt.Sprintf("cached:%s", *port)
	}

	fmt.Println(address)
	var deviceId int64 = 6248148189135751453
	poolCon := newClientConn(*connToOpen, address)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			//fmt.Println("Num of Pool", pool.GetNumberOfOpenConnections())
			deviceCount.Lock()
			fmt.Printf("Get RPC Count: %d -- TimeStamp: %d \n", deviceCount.currentCount, time.Now().Unix())
			fmt.Printf("Total RPC Count: %d \n", deviceCount.totalCount)
			fmt.Printf("Total Number of Connections in Use: %d \n", poolCon.getNumberOfConnectionsInUse())
			fmt.Printf("Total Number of Available Idle Connections: %d \n\n", poolCon.getNumberOfIdleConnections())
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
		//ticker.Stop()
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

func rpcCall(pool *connPool, deviceId int64, deviceCount *counter, i int) {
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
func makeCall(pool *connPool, deviceId int64) (*proto.Device, error) {
	//fmt.Println("call is being made")
	conn := pool.get()
	defer conn.release()
	cachedService := proto.NewCacheClient(conn.conn)
	deviceInput := &proto.DeviceInput{
		Id: deviceId,
	}
	return cachedService.GetDevice(context.Background(), deviceInput)
}

type conn struct {
	pool *connPool
	conn *grpc.ClientConn
}

type connPool struct {
	sync.Mutex
	connInUse     int
	availableConn []*conn
}

func newClientConn(num int, address string) *connPool {
	connPool := &connPool{}
	for i := num; i > 0; i-- {
		//fmt.Println(num)
		duration := time.Second * 5
		//timeout := time.After()
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		newConn, err := grpc.DialContext(ctx, address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())
		cancel()
		if err != nil {
			panic(err)
		}

		connPool.availableConn = append(connPool.availableConn, &conn{
			pool: connPool,
			conn: newConn,
		})
	}
	return connPool
}

func (cp *connPool) getNumberOfConnectionsInUse() int {
	cp.Lock()
	defer cp.Unlock()
	return cp.connInUse
}
func (cp *connPool) getNumberOfIdleConnections() int {
	cp.Lock()
	defer cp.Unlock()
	return len(cp.availableConn)
}
func (cp *connPool) get() *conn {
	for {
		cp.Lock()
		if len(cp.availableConn) > 0 {
			//fmt.Println("connection")
			conn := cp.availableConn[0]
			cp.availableConn = cp.availableConn[1:]
			cp.connInUse++
			cp.Unlock()
			return conn
		} else {
			cp.Unlock()
		}
	}
}

func (c *conn) release() {
	c.pool.Lock()
	c.pool.connInUse--
	c.pool.availableConn = append(c.pool.availableConn, c)
	c.pool.Unlock()
}
