package main

import (
	// "encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"io/ioutil"

	"github.com/olekukonko/tablewriter"
)

var filename string

var rootCmd = &cobra.Command{
	Use:   "ffbeauty",
	Short: "Show ffprobe like a beauty",
	Run:   cmdrun,
}

func cmdrun(cmd *cobra.Command, args []string) {

	var err error
	var data []byte
	if filename == "" {
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = ioutil.ReadFile(filename)
	}

	if err != nil {
		fmt.Println("Open file failed, detail:", err.Error())
		os.Exit(0)
	}

	proResp := FProbe{}
	if err := json.Unmarshal(data, &proResp); err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(80)
	// table.SetBorder(false)
	headers := []string{"PTS", "DTS"}
	table.SetHeader(headers)

	defer func() {
		// table.Render() // Send output
	}()

	aCnt := 0
	bpCnt := 0
	lastPts := 0
	size := 0
	finalPts := 0
	for _, f := range proResp.Frames {
		finalPts = int(f.PktDts / 1000)
		// a
		if f.KeyFrame == 1 && f.MediaType == "audio" {
			fmt.Print("A")
			aCnt += 1
		} else if f.KeyFrame == 1 && f.MediaType == "video" {
			// I
			if lastPts == 0 {
				fmt.Printf("\nBP:%d\tA:%d\tI pts:%d\t diff:%d\t",
					bpCnt, aCnt, f.PktDts/1000, 0)
			} else {
				elapsed := int(f.PktDts/1000) - lastPts
				fmt.Printf("\nBP:%d\tA:%d\tI pts:%d\tdiff:%d\tfps:%d\tbr:%dkbps\tsize:%dB\t",
					bpCnt, aCnt, f.PktDts/1000, int(f.PktDts/1000)-lastPts, (bpCnt+1)/elapsed, size/elapsed*8/1000, size)
			}
			lastPts = int(f.PktDts / 1000)
			aCnt = 0
			bpCnt = 0
			size = 0
		} else if f.KeyFrame == 0 && f.MediaType == "video" {
			// P
			if f.PictType == "B" {
				bpCnt += 1
				fmt.Print("B")
			} else if f.PictType == "P" {
				bpCnt += 1
				fmt.Print("P")
			} else if f.PictType == "I" {
				bpCnt += 1
				fmt.Print("i")
			} else {
				fmt.Print(f.PictType)
				panic("err")
			}

			if s, err := strconv.Atoi(f.PktSize); err != nil {
				panic(err)
			} else {
				size = size + s
			}
		} else {
			spew.Dump(f)
			panic("frame to process")
		}
	}

	elapsed := finalPts - lastPts
	fmt.Printf("\nBP:%d\tA:%d\tI pts:%d\tdiff:%d\tfps:%d\tbr:%dkbps\tsize:%dB",
		bpCnt, aCnt, finalPts, finalPts-lastPts, (bpCnt+1)/elapsed, size/elapsed*8/1000, size)

	// table.Append(line)

	return

}

func setupCmd() {
	rootCmd.PersistentFlags().StringVarP(&filename, "file", "f", "", "flv file, if do not set file then read from stdin")
	// rootCmd.PersistentFlags().BoolVar(&show_sei, "sei", false, "show sei info")
	// rootCmd.PersistentFlags().BoolVar(&show_only_nalt, "simple", false, "only show nal type")
	// rootCmd.PersistentFlags().BoolVar(&show_a, "a", false, "show audio")
	// rootCmd.PersistentFlags().BoolVar(&show_v, "v", true, "show video")
	// rootCmd.PersistentFlags().BoolVar(&no_show_i, "non-key", false, "use with -v:  do not show keyframes")
	rootCmd.MarkFlagRequired("file")
}

func main() {
	setupCmd()
	rootCmd.Execute()
}

type FProbe struct {
	Frames []struct {
		BestEffortTimestamp     int64  `json:"best_effort_timestamp"`
		BestEffortTimestampTime string `json:"best_effort_timestamp_time"`
		ChannelLayout           string `json:"channel_layout"`
		Channels                int64  `json:"channels"`
		ChromaLocation          string `json:"chroma_location"`
		CodedPictureNumber      int64  `json:"coded_picture_number"`
		DisplayPictureNumber    int64  `json:"display_picture_number"`
		Height                  int64  `json:"height"`
		InterlacedFrame         int64  `json:"interlaced_frame"`
		KeyFrame                int64  `json:"key_frame"`
		MediaType               string `json:"media_type"`
		NbSamples               int64  `json:"nb_samples"`
		PictType                string `json:"pict_type"`
		PixFmt                  string `json:"pix_fmt"`
		PktDts                  int64  `json:"pkt_dts"`
		PktDtsTime              string `json:"pkt_dts_time"`
		PktDuration             int64  `json:"pkt_duration"`
		PktDurationTime         string `json:"pkt_duration_time"`
		PktPos                  string `json:"pkt_pos"`
		PktPts                  int64  `json:"pkt_pts"`
		PktPtsTime              string `json:"pkt_pts_time"`
		PktSize                 string `json:"pkt_size"`
		RepeatPict              int64  `json:"repeat_pict"`
		SampleAspectRatio       string `json:"sample_aspect_ratio"`
		SampleFmt               string `json:"sample_fmt"`
		StreamIndex             int64  `json:"stream_index"`
		TopFieldFirst           int64  `json:"top_field_first"`
		Width                   int64  `json:"width"`
	} `json:"frames"`
}
