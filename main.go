package main

/*

#cgo CFLAGS: -I"C:\Program Files (x86)\OpenAL 1.1 SDK\include"


#cgo windows,386 LDFLAGS: C:\Program Files (x86)\OpenAL 1.1 SDK\libs\Win32\OpenAL32.lib
#cgo windows,amd64 LDFLAGS: C:\Program Files (x86)\OpenAL 1.1 SDK\libs\Win64\OpenAL32.lib

#cgo linux LDFLAGS: -lopenal

*/
import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

// SourceReader ...
type SourceReader struct {
	fromstring bool
	stringData string
	filePtr    *os.File
	myReader   io.Reader
	myBuffer   *bufio.Reader
}

// InitFromFile ...
func (r *SourceReader) InitFromFile(s string) {
	var err error
	r.filePtr, err = os.Open(s)
	if err != nil {
		log.Println("File unopenable.")
		log.Fatal(err)
	}
	r.fromstring = false
	r.Reset()
}

// InitFromString ...
func (r *SourceReader) InitFromString(s string) {
	r.fromstring = true
	r.stringData = s
	r.Reset()
}

// Reset ...
func (r *SourceReader) Reset() {
	if r.fromstring {
		r.myReader = strings.NewReader(r.stringData)
		r.myBuffer = bufio.NewReader(r.myReader)
	} else {
		r.myReader = io.Reader(r.filePtr)
		r.myBuffer = bufio.NewReader(r.myReader)
	}
	var err error
	_, err = r.myBuffer.ReadByte()
	if err != nil {
		log.Println("Data unreadble.")
		log.Fatal(err)
	}
	err = r.myBuffer.UnreadByte()
	if err != nil {
		log.Println("Data unreadble.")
		log.Fatal(err)
	}
}

// CloseFile ...
func (r *SourceReader) CloseFile() {
	if r.fromstring {
		r.filePtr.Close()
	}
}

func main() {
	// Args parse
	var filePath string
	flag.StringVar(&filePath, "f", "", "A file path")

	var stringData string
	flag.StringVar(&stringData, "s", "", "A string")

	flag.Parse()

	// vars
	var datas [2]byte
	var source SourceReader
	var err error

	// data buffer init
	if stringData == "" {
		if filePath == "" {
			filePath, err = os.Executable()
			if err != nil {
				log.Println("Executable path unknow.")
				log.Fatal(err)
			}
		}
		log.Print(filePath)
		source.InitFromFile(filePath)
	} else {
		source.InitFromString(stringData)
	}

	// Sound init
	master := snd.NewMixer()
	const buffers = 1
	if al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}
	defer al.CloseDevice()
	al.Start(master)
	defer al.Stop()

	sine := snd.Sine()
	square := snd.Square()
	mod := snd.NewOscil(sine, 0, nil)
	osc := snd.NewOscil(square, 440, mod) // oscillator

	master.Append(osc)
	al.Notify()

	// main loop
	var freqs [2]float64

	for range time.Tick(time.Second) {
		datas[0] = 0
		for datas[0] == 0 {
			if source.myBuffer.Buffered() == 0 {
				source.Reset()
				log.Println("Buffer reseted")
			}
			datas[0], err = source.myBuffer.ReadByte()
			if err != nil {
				log.Println("Source data unreadable.")
				log.Fatal(err)
			}
		}
		datas[1] = 0
		for datas[1] == 0 {
			if source.myBuffer.Buffered() == 0 {
				source.Reset()
				log.Println("Buffer reseted")
			}
			datas[1], err = source.myBuffer.ReadByte()
			if err != nil {
				log.Println("Source data unreadable.")
				log.Fatal(err)
			}
		}
		freqs[0] = 0.5 + float64(int(datas[0]))/64
		freqs[1] = 100 * (1 + float64(int(datas[1]))/16)
		log.Printf("> underruns=%-4v buflen=%-4v tickavg=%-12s drift=%-12s | bytes=%-4v freqs=%-9v\n",
			al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox(), datas, freqs)
		mod.SetFreq(freqs[0], nil)
		osc.SetFreq(freqs[1], mod)
	}
}
