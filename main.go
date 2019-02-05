package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

// OSStat is the obj type
type OSStat struct {
	Title               string
	TotalMemory         string
	FreeMemory          string
	PercentageMemory    string
	TotalDiskSpace      string
	FreeDiskSpace       string
	PercentageDiskSpace string
	HostName            string
	UpTime              string
	Process             string
	OS                  string
	Platform            string
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Root path
	r.GET("/", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, getOsStat())
	})

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.String(http.StatusOK, "pong")
	})

	return r
}

func getOsStat() OSStat {
	var osStat OSStat
	vmMemoryStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")
	hostStat, _ := host.Info()
	osStat.Title = "Welcome to a very light & simple GO Webserver! "
	osStat.TotalMemory = strconv.FormatUint(vmMemoryStat.Total, 10)
	osStat.FreeMemory = strconv.FormatUint(vmMemoryStat.Free, 10)
	osStat.PercentageMemory = strconv.FormatFloat(vmMemoryStat.UsedPercent, 'f', 2, 64)
	osStat.TotalDiskSpace = strconv.FormatUint(diskStat.Total, 10)
	osStat.FreeDiskSpace = strconv.FormatUint(diskStat.Free, 10)
	osStat.PercentageDiskSpace = strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64)
	osStat.HostName = hostStat.Hostname
	osStat.UpTime = strconv.FormatUint(hostStat.Uptime, 10)
	osStat.Process = strconv.FormatUint(hostStat.Procs, 10)
	osStat.OS = hostStat.OS
	osStat.Platform = hostStat.Platform
	return osStat
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
