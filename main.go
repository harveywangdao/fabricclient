package main

import (
	"fabricclient/fabric"
	"fabricclient/logger"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"sync"
)

const (
	logDirPath         = "log"
	logFilePath        = "log/fabric.log"
	FabricConfFilePath = "conf/my.ini"
)

func initLogger() error {
	/*	st, err := os.Stat(logDirPath)
		if err == nil {
			if !st.IsDir() {
				log.Fatal(logDirPath, "is not dir")
			}
		} else {
			if os.IsNotExist(err) {
				err = os.Mkdir(logDirPath, os.ModePerm)
				if err != nil {
					log.Fatal("mkdir fail")
				}
			} else {
				log.Fatal(logDirPath, "error")
			}
		}

		fileHandler := logger.NewFileHandler(logFilePath)
		logger.SetHandlers(logger.Console, fileHandler)
	*/
	logger.SetHandlers(logger.Console)
	//defer logger.Close()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)

	return nil
}

func main() {
	err := initLogger()
	if err != nil {
		log.Fatalln(err)
	}

	cfg, err := ini.Load(FabricConfFilePath)
	if err != nil {
		logger.Error(err)
		return
	}

	ipport := cfg.Section("").Key("FabricServerIpPort").String()

	var wg sync.WaitGroup
	wg.Add(1)

	_, err = fabric.NewFabricClient(ipport, &wg)
	if err != nil {
		logger.Error(err)
		return
	}

	wg.Wait()
	logger.Debug("fabric client exit")
}
