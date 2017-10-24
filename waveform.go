package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

	// TODO: parse wav header instead of just skip it
	data = data[wavHeaderSize:]
	for i := 0; i < len(data)-1; i += 2 {
		sample := binary.LittleEndian.Uint16(data[i : i+2])
		audio = append(audio, int16(sample))
	}

	return audio
}

func plotAudio(imgfile string, audio []int16, samplerate int) {
	f, err := ioutil.TempFile("", "waveform")
	abortonerr(err, "creating tmp file to generate waveform")
	defer f.Close()

	sampleratePerMili := float32(samplerate) / 1000

	for i, sample := range audio {
		timestamp := float32(i) / sampleratePerMili
		f.WriteString(fmt.Sprintf("%f %d\n", timestamp, sample))
	}

	fmt.Printf("tmp gnuplot file: [%s]\n", f.Name())

	gnuplotscript := []string{
		"set terminal svg",
		fmt.Sprintf("set output '%s'", imgfile),
		fmt.Sprintf("plot '%s' every 35 with lines", f.Name()),
		`set xlabel "time (ms)"`,
		`set ylabel "sample value (signed int)`,
		"set bmargin 0",
	}

	gnuplotargs := strings.Join(gnuplotscript, ";")
	fmt.Printf("running gnuplot: [%s]\n", gnuplotargs)
	cmd := exec.Command("gnuplot", "-e", gnuplotargs)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	abortonerr(err, "running gnuplot")

	//TODO: delete tmp file
}

func main() {
	const samplerate = 8000
	var audiofile string

	flag.StringVar(&audiofile, "audio", "", "path to audio file (only WAV PCM 8000Hz LE supported)")

	flag.Parse()

	if audiofile == "" {
		flag.Usage()
		return
	}

	audio := loadAudio(audiofile)
	plotAudio(audiofile+".waveform.svg", audio, samplerate)

	fmt.Println("done")
}
