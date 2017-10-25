package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func abortonerr(err error, op string) {
	if err != nil {
		fmt.Printf("%s: fatal error: %s\n", op, err)
		os.Exit(1)
	}
}

func loadAudio(audiofile string) (*wavHeader, []int16) {
	// TODO(i4k): calculate the offset of data chunk
	const wavHeaderSize = 44

	audio := []int16{}
	f, err := os.Open(audiofile)
	abortonerr(err, "opening audio file")
	defer f.Close()

	hdr, err := parseHeader(f)
	abortonerr(err, "parsing WAV header")

	// header already skipped

	data, err := ioutil.ReadAll(f)
	abortonerr(err, "opening audio file")

	for i := 0; i < len(data)-1; i += 2 {
		sample := binary.LittleEndian.Uint16(data[i : i+2])
		audio = append(audio, int16(sample))
	}

	return hdr, audio
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
		hdr, audio := loadAudio(audiofile)

		plotAudio(output, audio, hdr.RIFFChunkFmt.SampleRate)
		return
	}

	if audiofiles != "" {
		parsedAudios := strings.Split(audiofiles, ",")
		audios := [][]int16{}
		audionames := []string{}

		var (
			hdr  *wavHeader
			data []int16
		)
		for _, parsedAudiofile := range parsedAudios {
			fmt.Printf("loading audio[%s]\n", parsedAudiofile)
			hdr, data = loadAudio(parsedAudiofile)
			audios = append(audios, data)
			audionames = append(audionames, filepath.Base(parsedAudiofile))
		}
		plotAudios(output, audios, audionames, hdr.RIFFChunkFmt.SampleRate)
	}
}
