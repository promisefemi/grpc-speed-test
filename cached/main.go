package main

import (
	"cached/model"
	"cached/proto"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	pj "google.golang.org/protobuf/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

func unPackToProto(in interface{}, out pj.Message) error {

	//	Marshal in into byte
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(b, out); err != nil {
		return err
	}
	return nil
}

type service struct {
	proto.UnimplementedCacheServer
	cache *model.DeviceCache
}
type counter struct {
	sync.Mutex
	entries                  []int
	currentCount, totalCount int
}

var deviceCount counter

func (s service) GetDevice(ctx context.Context, deviceInput *proto.DeviceInput) (*proto.Device, error) {
	protoDevice := new(proto.Device)

	device, ok := s.cache.Data[deviceInput.Id]
	if !ok {
		return nil, fmt.Errorf("could not find device in cache")
	}
	err := unPackToProto(device, protoDevice)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	deviceCount.Lock()
	deviceCount.currentCount++
	deviceCount.totalCount++
	deviceCount.Unlock()
	return protoDevice, nil
}

func newService(cache *model.DeviceCache) *service {
	return &service{
		cache: cache,
	}
}

func main() {

	port := flag.String("port", "9000", "Port")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		panic(err)
	}

	devicesCache := fillUpCache("devices.json")

	cachedService := newService(devicesCache)
	grpcServer := grpc.NewServer()
	proto.RegisterCacheServer(grpcServer, cachedService)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			//fmt.Println("Num of Pool", pool.GetNumberOfOpenConnections())
			deviceCount.Lock()
			fmt.Printf("Get RPC Count: %d -- TimeStamp: %d \n", deviceCount.currentCount, time.Now().Unix())
			fmt.Printf("Total RPC Count: %d \n\n", deviceCount.totalCount)
			deviceCount.entries = append(deviceCount.entries, deviceCount.currentCount)
			deviceCount.currentCount = 0
			deviceCount.Unlock()
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		ticker.Stop()
		total := 0
		for _, entry := range deviceCount.entries {
			total += entry
		}
		fmt.Printf("\nTotal request handled %d\n", deviceCount.totalCount)
		fmt.Printf("Average request made per second %d\n", total/len(deviceCount.entries))
		os.Exit(0)
	}()

	log.Printf("Server is running on port %s", *port)
	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}
}

func fillUpCache(fileDir string) *model.DeviceCache {

	log.Println("setting up local cache")
	Rows := make([][]interface{}, 0)

	absPath, err := filepath.Abs(fileDir)
	if err != nil {
		panic(err)
	}

	fileByte, err := os.ReadFile(absPath)
	if err != nil {
		panic(err)
	}
	log.Println("loading cache from json file ")
	d := json.NewDecoder(strings.NewReader(string(fileByte)))
	d.UseNumber()
	err = d.Decode(&Rows)
	if err != nil {
		panic(err)
	}

	timeLayout := "2006-01-02 15:04:05"
	var devices = new(model.DeviceCache)
	devices.Data = make(map[int64]*model.Device)
	_ = devices
	_ = timeLayout

	for _, row := range Rows {
		var device model.Device

		val, err := row[0].(json.Number).Int64()
		handleError(err)
		device.Id = val

		//Type
		val, err = row[1].(json.Number).Int64()
		handleError(err)
		device.Type.Set(int32(val))

		device.HardwareId = row[4].(string)
		device.FirmwareId = row[5].(string)

		//USer ID
		val, err = row[7].(json.Number).Int64()
		handleError(err)
		device.UserId = val

		//Bridge ID
		mainVal, ok := row[8].(json.Number)
		if ok {
			val, err = mainVal.Int64()
			handleError(err)
			device.BridgeId.Set(val)
		}
		//SampleMultiplier
		valFloat, err := row[9].(json.Number).Float64()
		handleError(err)
		device.SampleMultiplier = valFloat

		device.Settings = row[10].(string)

		//TsSchemaVer
		val, err = row[13].(json.Number).Int64()
		handleError(err)
		device.TsSchemaVer = val

		//Added Last Seen
		nt, err := time.Parse(timeLayout, row[14].(string))
		if err != nil {
			panic(err)
		}
		device.LastSeen.Set(nt)
		//Last node
		device.LastNode = row[15].(string)

		//Last Sample Value
		mainVal, ok = row[16].(json.Number)
		if ok {
			val, err = mainVal.Int64()
			handleError(err)
			device.LastSampleValue.Set(val)
		}

		//Last Sample Date time
		nt, _ = time.Parse(timeLayout, row[17].(string))
		if !nt.IsZero() {
			device.LastSampleDateTime.Set(nt)
		}

		mainVal, ok = row[18].(json.Number)
		if ok {
			val, err = mainVal.Int64()
			handleError(err)
			device.Connected = int32(val)
		}
		device.LastCacheTimestamp = time.Now().Unix()
		devices.Data[device.Id] = &device
		devices.Count++

	}

	return devices
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
