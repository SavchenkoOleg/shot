package conf

import "os"

type FlagConfigStruct struct {
	ServerAdress    string
	BaseURL         string
	FileStoragePath string
}

type ServConfigtruct struct {
	NewURLPref      string
	ServerAdress    string
	BaseURL         string
	FullPathTest    string
	FileStorage     bool
	FileStoragePath string
}

var ServConfig ServConfigtruct
var FlagConfig FlagConfigStruct

func HendlerSetting() (outConf ServConfigtruct) {

	// значения по умолчанию
	outConf.NewURLPref = "newURL"
	outConf.ServerAdress = "localhost:8080"
	outConf.BaseURL = "shot"
	outConf.FileStorage = false
	outConf.FileStoragePath = ""

	// переменные окружения
	BaseURL, exp := os.LookupEnv("BASE_URL")
	if exp {
		outConf.BaseURL = BaseURL
	}

	serverAdress, exp := os.LookupEnv("SERVER_ADDRESS")
	if exp {
		outConf.ServerAdress = serverAdress
	}

	outConf.FileStoragePath, outConf.FileStorage = os.LookupEnv("FILE_STORAGE_PATH")

	// флаги, если они есть
	if FlagConfig.BaseURL != "" {
		outConf.BaseURL = FlagConfig.BaseURL

	}

	if FlagConfig.ServerAdress != "" {
		outConf.ServerAdress = FlagConfig.ServerAdress

	}

	if FlagConfig.FileStoragePath != "" {
		outConf.FileStoragePath = FlagConfig.FileStoragePath
		outConf.FileStorage = true
	}

	//  синтез пути, используемый в тестах
	outConf.FullPathTest = "http://" + outConf.ServerAdress + "/" + outConf.BaseURL + "/" + outConf.NewURLPref

	return outConf
}
