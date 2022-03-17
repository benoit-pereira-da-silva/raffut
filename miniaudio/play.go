package miniaudio

import (
	"fmt"
	"github.com/hajimehoshi/go-mp3"
	"github.com/youpy/go-riff"
	"github.com/youpy/go-wav"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Play a local or distant mp3 and wave files to test the audio layer.
func (p *Miniaudio) Play(access string, echo bool) error {
	p.Configure(access, 0, 0, echo, nil)
	var input io.ReadCloser
	var decoder io.Reader
	if strings.Contains(access, "http://") || strings.Contains(access, "https://") {
		resp, err := http.Get(access)
		if err != nil {
			return err
		}
		input = resp.Body
	} else {
		file, err := os.Open(access)
		if err != nil {
			return err
		}
		input = file
		defer file.Close()
	}
	ext := filepath.Ext(access)
	switch ext {
	case ".mp3":
		mp3Dec, err := mp3.NewDecoder(input)
		if err != nil {
			return err
		}
		p.nbChannels = 2
		p.sampleRate = float64(mp3Dec.SampleRate())
		decoder = mp3Dec
	case ".wav":
		waveDec := wav.NewReader((input).(riff.RIFFReader))
		f, err := waveDec.Format()
		if err != nil {
			return err
		}
		p.nbChannels = int(f.NumChannels)
		p.sampleRate = float64(f.SampleRate)
		decoder = waveDec
	default:
		return fmt.Errorf("not a valid file")
	}
	return p.ReadStreamFrom(decoder)
}
