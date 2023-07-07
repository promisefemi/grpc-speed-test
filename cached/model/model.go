package model

import (
	"sync"
	"time"
)

//type DTime sql.NullTime
//
//func (dt *DTime) Scan(value interface{}) error {
//	fmt.Println(value.(string), "VALue ")
//	if value == nil {
//		return errors.New("invalid value")
//	}
//
//	return nilMessageType = 6
//}

type DeviceCache struct {
	Data  map[int64]*Device
	Count int
}

type FrameType int

const (
	FT_SAMPLE FrameType = iota + 1
	FT_STATE
	FT_LOG
	FT_EVENT
	FT_OTA
	FT_HEARTBEAT
	FT_BOOT
	FT_ORIENT
	FT_KEY_N
	FT_KEY_KK
	FT_KEY_XX
)

func (f FrameType) IsValid() bool {
	switch f {
	case FT_SAMPLE,
		FT_STATE,
		FT_LOG,
		FT_EVENT,
		FT_OTA,
		FT_HEARTBEAT,
		FT_BOOT,
		FT_ORIENT,
		FT_KEY_N,
		FT_KEY_KK,
		FT_KEY_XX:
		return true
	default:
		return false
	}
}

func (f FrameType) String() string {
	return []string{
		"FT_SAMPLE",
		"FT_STATE",
		"FT_LOG",
		"FT_EVENT",
		"FT_OTA",
		"FT_HEARTBEAT",
		"FT_BOOT",
		"FT_ORIENT",
		"FT_KEY_N",
		"FT_KEY_KK",
		"FT_KEY_XX",
	}[f-1]
}

type DeviceInput struct {
	Id        int64     `json:"Id,omitempty"`
	FrameType FrameType `json:"FrameType,omitempty"`
}

type Device struct {
	Mu sync.Mutex `protojson:"-" json:"-"`
	//Read only fields
	Id               int64     `json:"Id,omitempty"`
	Type             NullInt32 `json:"Type,omitempty"`
	HardwareId       string    `json:"HardwareId,omitempty"`
	FirmwareId       string    `json:"FirmwareId,omitempty"`
	BridgeId         NullInt64 `json:"BridgeId,omitempty"`
	SampleMultiplier float64   `json:"SampleMultiplier,omitempty"`
	Settings         string    `json:"Settings,omitempty"`
	Timezone         string    `json:"Timezone"`
	UserId           int64     `json:"UserId"`
	TsSchemaVer      int64     `json:"TsSchemaVer,omitempty"`
	Features         string    `json:"-"`
	KeyHex           string    `json:"-"`

	//update with all message types
	LastSeen  NullTime `json:"LastSeen,omitempty"`
	LastNode  string   `json:"LastNode,omitempty"`
	Connected int32    `json:"Connected"`

	//update for SAMPLE message types
	LastSampleValue    NullInt64 `json:"LastSampleValue,omitempty"`
	LastSampleDateTime NullTime  `json:"LastSampleDateTime,omitempty"`

	//update for CALIBRATE message types
	AxisCalibration string `json:"AxisCalibration,omitempty"`

	//device Rules
	DeviceRules map[int32]*DeviceRule `json:"DeviceRules,omitempty"`

	//device State
	LastState map[string]int32 `json:"LastState,omitempty"`

	//time since device was cached
	LastCacheTimestamp int64 `json:"-"`

	//Dirty is the flag that checks if a device has been synced with mysql
	Dirty bool `json:"-"`
}

//func (d *Device) Scan(src interface{}) error {

// return nil
// }
// func (d *Device) Value(src interface{}) error {
//
// }
type DeviceRules struct {
	Rules []*DeviceRule
}
type DeviceRule struct {
	//Readonly fields
	Id              int32         `json:"Id,omitempty"`
	Active          int32         `json:"Active,omitempty"`
	Type            int32         `json:"Type,omitempty"`
	DeviceId        int64         `json:"DeviceId,omitempty"`
	Bucket          int32         `json:"Bucket,omitempty"`
	BucketCount     int32         `json:"BucketCount,omitempty"`
	Property        NullString    `json:"Property,omitempty"`
	SinceDateTime   NullTime      `json:"SinceDateTime,omitempty"`
	UntilDateTime   NullTime      `json:"UntilDateTime,omitempty"`
	Comparison      Comparison    `json:"Comparison,omitempty"`
	GroupMultiplier int32         `json:"GroupMultiplier,omitempty"`
	Operation       OperationEnum `json:"Operation,omitempty"`
	Value           float32       `json:"Value,omitempty"`

	LastTrigger  NullTime `json:"LastTrigger,omitempty"`  //R/W
	NextEval     NullTime `json:"NextEval,omitempty"`     //R/W
	NextEvalSecs int32    `json:"NextEvalSecs,omitempty"` //R/W
}

type Comparison string

const (
	ComparisonGreaterThan        Comparison = ">"
	ComparisonLessThan           Comparison = "<"
	ComparisonGreaterThanOrEqual Comparison = ">="
	ComparisonLessThanOrEqual    Comparison = "<="
	ComparisonEqualTo            Comparison = "=="
	ComparisonNotEqual           Comparison = "!="
)

func (c Comparison) IsValid() bool {
	switch c {
	case ComparisonGreaterThan, ComparisonLessThan, ComparisonGreaterThanOrEqual,
		ComparisonLessThanOrEqual, ComparisonNotEqual, ComparisonEqualTo:
		return true
	default:
		return false
	}
}
func (c Comparison) String() string {
	return string(c)
}

type OperationEnum string

const (
	OperationalEnumSum OperationEnum = "sum"
	OperationalEnumMin OperationEnum = "min"
	OperationalEnumMax OperationEnum = "max"
	OperationalEnumAvg OperationEnum = "avg"
	OperationalEnumCnt OperationEnum = "cmt"
)

func (o OperationEnum) IsValid() bool {
	switch o {
	case OperationalEnumCnt, OperationalEnumMin, OperationalEnumMax, OperationalEnumAvg, OperationalEnumSum:
		return true
	default:
		return false
	}
}
func (o OperationEnum) String() string {
	return string(o)
}

type DeviceUpdate struct {
	Id        int64     `json:"Id,omitempty"`
	FrameType FrameType `json:"FrameType,omitempty"`
	LastSeen  int64     `json:"LastSeen,omitempty"`
	LastSeenT time.Time

	LastNode  string `json:"LastNode,omitempty"`
	Connected int32  `json:"Connected,omitempty"`

	//update for SAMPLE message types
	LastSampleValue     int64 `json:"LastSampleValue,omitempty"`
	LastSampleDateTime  int64 `json:"LastSampleDateTime,omitempty"`
	LastSampleDateTimeT time.Time

	// Update for calibration frameType
	AxisCalibration string `json:"AxisCalibration,omitempty"`

	//Update for boot frameType
	FirmwareId string `json:"FirmwareId,omitempty"`
	BridgeId   int64  `json:"BridgeId,omitempty"`
}

type DeviceRuleUpdate struct {
	Id       int32 `json:"Id,omitempty"`
	DeviceId int64 `json:"DeviceId,omitempty"`

	LastTrigger  int64     `json:"LastTrigger,omitempty"` //R/W
	LastTriggerT time.Time //R/W

	NextEval  int64 `json:"NextEval,omitempty"` //R/W
	NextEvalT time.Time

	NextEvalSecs int32 `json:"NextEvalSecs,omitempty"` //R/W
}
type DeviceState struct {
	Id     int64            `json:"Id,omitempty"`
	Type   int32            `json:"Type,omitempty"`
	States map[string]int32 `json:"States,omitempty"`
}
