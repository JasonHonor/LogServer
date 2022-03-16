package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/mcuadros/go-syslog.v2"
)

const (
	//LOGPATH  LOGPATH/time.Now().Format(FORMAT)/*.log
	LOGPATH = "log/"
	//FORMAT .
	FORMAT = "20060102"
	//LineFeed 换行
	LineFeed = "\n"
)

//CreateDir  文件夹创建
func CreateDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	os.Chmod(path, os.ModePerm)
	return nil
}

//IsExist  判断文件夹/文件是否存在  存在返回 true
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func WriteLog(path, fileName, msg string) error {
	if !IsExist(path) {
		CreateDir(path)
	}
	var (
		err error
		f   *os.File
	)

	f, err = os.OpenFile(path+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	_, err = io.WriteString(f, LineFeed+msg)

	defer f.Close()
	return err
}

func main() {
	fmt.Println("starting...")

	dir, errDir := filepath.Abs(filepath.Dir(os.Args[0]))
	if errDir != nil {
		fmt.Println(errDir)
		return
	}

	var logPath = dir + "/" + LOGPATH + "/"
	WriteLog(logPath, "server.log", "Starting...")
	WriteLog(logPath, "server.log", "CurrentDir "+dir)

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)

	err1 := server.ListenUDP("0.0.0.0:514")

	if err1 != nil {
		WriteLog(logPath, "server.log", err1.Error())
	}

	err2 := server.ListenTCP("0.0.0.0:514")
	if err2 != nil {
		WriteLog(logPath, "server.log", err2.Error())
	}

	server.Boot()

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			//fmt.Println(logParts)
			//map[client:192.168.61.254:514 content:UDPv4 WRITE [85] to [AF_INET]219.232.205.138:1194: P_DATA_V1 kid=4 DATA len=84
			//facility:3 hostname:nx2 priority:29 severity:5 tag:openvpn timestamp:2019-12-15 19:11:14 +0000 UTC tls_peer:]
			msg := fmt.Sprintf("%s|%s^%s^%s", time.Now().Format("2006-01-02T15:04:05"), logParts["timestamp"], logParts["tag"], logParts["content"])

			//拆分客户端信息logParts["client"]
			sClientInfo := strings.Split(logParts["client"].(string), ":")
			if len(sClientInfo) >= 1 {
				path := logPath + sClientInfo[0] + "/"
				WriteLog(path, time.Now().Format(FORMAT)+".log", msg)
			}
		}
	}(channel)

	server.Wait()
}
