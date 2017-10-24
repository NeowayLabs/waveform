package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func abortonerr(err error, op string) {
	if err != nil {
		fmt.Printf("%s: fatal error: %s\n", op, err)
		os.Exit(1)
	}
}

func loadAudio(audiofile string) []int16 {
	const wavHeaderSize = 44

	audio := []int16{}
	f, err := os.Open(audiofile)
	abortonerr(err, "opening audio file")
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	abortonerr(err, "opening audio file")

	data = data[wavHeaderSize:]
	for i := 0; i < len(data)-1; i += 2 {
		sample := binary.LittleEndian.Uint16(data[i : i+2])
		audio = append(audio, int16(sample))
	}

	return audio
}

func main() {
	var audiofile string
	flag.StringVar(&audiofile, "audio", "", "path to audio file (only WAV PCM 8000Hz LE supported)")

	flag.Parse()

	if audiofile == "" {
		flag.Usage()
		return
	}

	audio := loadAudio(audiofile)
	fmt.Println(audio)
}
