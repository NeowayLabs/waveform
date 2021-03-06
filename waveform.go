package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/NeowayLabs/waveparser"
)

func abortonerr(err error, op string) {
	if err != nil {
		fmt.Printf("%s: fatal error: %s\n", op, err)
		os.Exit(1)
	}
}

func plotAudio(imgfile string, audio []int16, samplerate uint32) {
	f, err := ioutil.TempFile("", "waveform")
	abortonerr(err, "creating tmp file to generate waveform")
	defer f.Close()

	sampleratePerMili := float32(samplerate) / 1000

	for i, sample := range audio {
		timestamp := float32(i) / sampleratePerMili
		f.WriteString(fmt.Sprintf("%f %d %d\n", timestamp, sample)) // missing format value here
	}

	gnuplotscript := []string{
		"set terminal svg",
		fmt.Sprintf("set output '%s'", imgfile),
		fmt.Sprintf("plot '%s' every 35 with lines", f.Name()),
		`set xlabel "time (ms)"`,
		`set ylabel "sample value (signed int)`,
	}

	gnuplot(gnuplotscript)

	//TODO: delete tmp file
}

func gnuplot(script []string) {
	gnuplotargs := strings.Join(script, ";")
	cmd := exec.Command("gnuplot", "-e", gnuplotargs)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	abortonerr(err, "running gnuplot")
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

func plotAudios(imgfile string, audios [][]int16, audionames []string, samplerate uint32) {
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

	gnuplotscript := []string{
		"set terminal svg",
		fmt.Sprintf("set output '%s'", imgfile),
		`set xlabel "time (ms)"`,
		`set ylabel "sample value (signed int)"`,
	}

	plots := []string{}
	for i, audioname := range audionames {
		plots = append(plots, fmt.Sprintf(
			`"%s" every 30 using 1:%d title "%s" with lines`,
			filepath.ToSlash(f.Name()),
			i+2,
			audioname,
		))
	}

	plot := fmt.Sprintf("plot %s", strings.Join(plots, ","))
	gnuplotscript = append(gnuplotscript, plot)

	gnuplot(gnuplotscript)

	//TODO: delete tmp file
}

func main() {
	var audiofile string
	var audiofiles string
	var output string

	flag.StringVar(&audiofile, "audio", "", "path to audio file")
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
		wav, err := waveparser.Load(audiofile)
		abortonerr(err, "loading audio")
		audio, err := wav.Int16LESamples()
		abortonerr(err, "getting audio samples")
		plotAudio(output, audio, wav.Header.RIFFChunkFmt.SampleRate)
		return
	}

	if audiofiles != "" {
		parsedAudios := strings.Split(audiofiles, ",")
		audios := [][]int16{}
		audionames := []string{}
		var samplerate uint32

		for _, parsedAudiofile := range parsedAudios {
			fmt.Printf("loading audio[%s]\n", parsedAudiofile)
			wav, err := waveparser.Load(parsedAudiofile)
			abortonerr(err, "loading audio")
			audio, err := wav.Int16LESamples()
			abortonerr(err, "getting samples")
			audios = append(audios, audio)
			audionames = append(audionames, filepath.Base(parsedAudiofile))
			// TODO: not good way to handle that
			samplerate = wav.Header.RIFFChunkFmt.SampleRate
		}

		plotAudios(output, audios, audionames, samplerate)
	}
}
