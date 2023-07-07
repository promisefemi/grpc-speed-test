package model

type Config struct {
	General struct {
		Env string `json:"env"`
	} `json:"general"`
	Daemon struct {
		Background bool   `json:"background"`        //if background false set log dest += STDOUT
		Port       int    `json:"port"`              //grpc listen port
		User       string `jsonAutoGenerated:"user"` //probably won't be needed
		PidFile    string `json:"pid_file"`
	} `json:"daemon"`
	Cache struct {
		Devices struct {
			ExpireSecs int `json:"expire_secs"`
			DirtySecs  int `json:"dirty_secs"`
		} `json:"devices"`
		Constants struct {
			ExpireSecs int `json:"expire_secs"`
		} `json:"constants"`
		Rules struct {
			ExpireSecs int `json:"expire_secs"`
		} `json:"rules"`
	} `json:"cache"`
	Mysql struct {
		Dsn             string `json:"dsn"`
		MultiStatements bool   `json:"multiStatements"`
		ParseTime       bool   `json:"parseTime"`
		Timeout         string `json:"timeout"`
		MaxOpenConns    int    `json:"MaxOpenConns"`
		MaxConnLifetime int    `json:"MaxConnLifetime"`
	} `json:"mysql"`
	Logging struct {
		Dir           string `json:"dir"`
		File          string `json:"file"`
		Level         int    `json:"level"`
		Format        string `json:"format"`
		Usr1DebugSecs int    `json:"usr1_debug_secs"`
	} `json:"logging"`
}
