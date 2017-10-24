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
		f.WriteString(fmt.Sprintf("%f %d %d\n", timestamp, sample))
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

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func minAudioSize(audios [][]int16) int {
	m := len(audios[0])
	audios = audios[1:]
	for _, audio := range audios {
		m = min(m, len(audio))
	}
	return m
}

func plotAudios(imgfile string, audios [][]int16, samplerate int) {
	f, err := ioutil.TempFile("", "waveform")
	abortonerr(err, "creating tmp file to generate waveform")
	defer f.Close()

	sampleratePerMili := float32(samplerate) / 1000

	minSize := minAudioSize(audios)

	for i := 0; i < minSize; i++ {
		timestamp := float32(i) / sampleratePerMili
		f.WriteString(fmt.Sprintf("%f", timestamp))
		for _, audio := range audios {
			f.WriteString(fmt.Sprintf(" %d", audio[i]))
		}
		f.WriteString("\n")
	}

	fmt.Printf("tmp gnuplot file: [%s]\n", f.Name())

	//gnuplotscript := []string{
	//"set terminal svg",
	//fmt.Sprintf("set output '%s'", imgfile),
	//fmt.Sprintf("plot '%s' every 35 with lines", f.Name()),
	//`set xlabel "time (ms)"`,
	//`set ylabel "sample value (signed int)`,
	//"set bmargin 0",
	//}

	//gnuplotargs := strings.Join(gnuplotscript, ";")
	//fmt.Printf("running gnuplot: [%s]\n", gnuplotargs)
	//cmd := exec.Command("gnuplot", "-e", gnuplotargs)
	//cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout
	//err = cmd.Run()
	//abortonerr(err, "running gnuplot")

	//TODO: delete tmp file
}

func main() {
	const samplerate = 8000
	var audiofile string
	var audiofiles string
	var output string

	flag.StringVar(&audiofile, "audio", "", "path to audio file (only WAV PCM 8000Hz LE supported)")
	flag.StringVar(&audiofiles, "audios", "", "comma separated list of audio files to compare")
	flag.StringVar(&output, "output", "", "where the generated waveform will be saved")

	flag.Parse()

	if output == "" {
		flag.Usage()
		return
	}

	if audiofile == "" && audiofiles == "" {
		flag.Usage()
		return
	}

	if audiofile != "" {
		fmt.Printf("generating audio[%s] waveform[%s]\n", audiofile, output)
		audio := loadAudio(audiofile)
		plotAudio(output, audio, samplerate)
		return
	}

	if audiofiles != "" {
		parsedAudios := strings.Split(audiofiles, ",")
		audios := [][]int16{}

		for _, parsedAudiofile := range parsedAudios {
			fmt.Printf("loading audio[%s]\n", parsedAudiofile)
			audios = append(audios, loadAudio(parsedAudiofile))
		}
		plotAudios(output, audios, samplerate)
	}
}
