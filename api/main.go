package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"os"
	"os/exec"
	"souv/utils"
	"strconv"
	"time"
)

type RunningFFMPEGTask struct {
	pid       int
	time      time.Time
	channelId string
}

type StartStreamBody struct {
	RtspUrl   string `json:"rtspUrl"`
	ChannelId string `json:"channelId"`
}

var globalPidList []RunningFFMPEGTask

func startRTSP(rtspURL string, channelId string) {
	ffmpegArgs := []string{"-fflags", "nobuffer",
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-vsync", "0",
		"-copyts",
		"-vcodec", "copy",
		"-movflags", "frag_keyframe+empty_moov",
		"-an",
		"-hls_flags", "delete_segments+append_list",
		"-f", "segment",
		"-segment_list_flags", "live",
		"-segment_time", "1",
		"-segment_list_size", "3",
		"-segment_format", "mpegts",
		"-segment_list", "index.m3u8",
		"-segment_list_type", "m3u8",
		//"-segment_list_entry_prefix", channelId + "/",
		"-segment_wrap", "10",
		"%d.ts"}

	// Create directory for HLS files
	exec.Command("mkdir", "./stream/"+channelId).Run()

	cmd := exec.Command("ffmpeg", ffmpegArgs...)
	cmd.Dir = "./stream/" + channelId

	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Started streaming at " + channelId)

	// Add to global list of running agents
	globalPidList = append(globalPidList,
		RunningFFMPEGTask{cmd.Process.Pid, time.Now(), channelId},
	)
}

func killAgent(task *RunningFFMPEGTask) {
	process, _ := os.FindProcess(task.pid)

	err := process.Kill()
	if err != nil {
		fmt.Println("Failed to kill FFMPEG agent at pid" + strconv.Itoa(task.pid))
		fmt.Println(err)
	} else {
		fmt.Println("Killed agent at PID " + strconv.Itoa(task.pid))
		err := exec.Command("rm", "-rf", "./stream/"+task.channelId).Run()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func cleanupTasks() {
	// Get list of all classes to clean up and kill them
	var taskIndexesToRemove []int
	for i, task := range globalPidList {
		// More than 30 seconds ago
		if task.time.Before(time.Now().Add(-30 * time.Second)) {
			taskIndexesToRemove = append(taskIndexesToRemove, i)
			killAgent(&task)
		}
	}

	// Remove cancelled tasks from global list
	var newRunningPids []RunningFFMPEGTask
	for i, task := range globalPidList {
		if !utils.IntInSlice(i, taskIndexesToRemove) {
			newRunningPids = append(newRunningPids, task)
		}
	}
	globalPidList = newRunningPids
}

func startStreamController(c *gin.Context) {
	var body StartStreamBody
	err := c.BindJSON(&body)
	if err != nil {
		fmt.Println("Unable to parse body")
		return
	}

	// Check if task with this channel ID is already running
	var running = false
	var runningIndex = -1
	for i, task := range globalPidList {
		if task.channelId == body.ChannelId {
			running = true
			runningIndex = i
		}
	}

	if running {
		// Update keep alive time of existing stream
		globalPidList[runningIndex] = RunningFFMPEGTask{
			globalPidList[runningIndex].pid,
			time.Now(),
			globalPidList[runningIndex].channelId,
		}
	} else {
		// Spin up new ffmpeg task
		startRTSP(body.RtspUrl, body.ChannelId)
	}
}

func main() {

	// Periodic FFMPEG agent cleanup task
	go func() {
		for true {
			fmt.Println("Current running agents: " + strconv.Itoa(len(globalPidList)))
			fmt.Println("Running periodic agent cleanup task")
			cleanupTasks()
			time.Sleep(10 * time.Second) // Run every 10 seconds
		}
	}()

	exec.Command("rm", "-rf", "./stream").Run()
	exec.Command("mkdir", "./stream").Run()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("./stream", "./stream")
	router.POST("api/v1/stream", startStreamController)
	router.Run("0.0.0.0:8080")
}
