package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func startRTSP(rtspURL string, dirName string) {
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
		"-segment_list_entry_prefix", dirName,
		"-segment_wrap", "10",
		"%d.ts"}

	exec.Command("mkdir", "."+dirName).Run()

	cmd := exec.Command("ffmpeg", ffmpegArgs...)
	cmd.Dir = "." + dirName
	err := cmd.Start()

	fmt.Println(err)
	fmt.Println("Started streaming at " + dirName)
}

func startAgents() {
	exec.Command("rm", "-rf", "./stream").Run()
	exec.Command("mkdir", "./stream").Run()

	env, _ := os.LookupEnv("SOUV_RTSP_URLS")
	urls := strings.Split(env, ",")

	for i, url := range urls {
		startRTSP(url, "/stream/ch"+strconv.Itoa(i)+"/")
	}
}

func main() {
	router := gin.Default()
	startAgents()
	router.Static("./stream", "./stream")
	router.Run("0.0.0.0:8080")
}
