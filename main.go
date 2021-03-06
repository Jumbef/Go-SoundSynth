package main

/*

#cgo CFLAGS: -I"C:\Program Files (x86)\OpenAL 1.1 SDK\include"


#cgo windows,386 LDFLAGS: C:\Program Files (x86)\OpenAL 1.1 SDK\libs\Win32\OpenAL32.lib
#cgo windows,amd64 LDFLAGS: C:\Program Files (x86)\OpenAL 1.1 SDK\libs\Win64\OpenAL32.lib

#cgo linux LDFLAGS: -lopenal

*/
import (
	"bufio"
	"log"
	"os"
	"time"

	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

func main() {

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(ex)

	f, err := os.Open(ex)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	master := snd.NewMixer()
	const buffers = 1
	if al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}
	al.Start(master)

	sine := snd.Sine()
	mod := snd.NewOscil(sine, 0, nil)
	osc := snd.NewOscil(sine, 440, mod) // oscillator

	master.Append(osc)
	al.Notify()

	freq := float64(1)
	data := byte(0)
	for range time.Tick(time.Second) {
		data = 0
		for data == 0 {
			data, err = reader.ReadByte()
			if err != nil {
				log.Fatal(err)
			}
		}
		freq = 0.5 + float64(int(data))/64
		log.Printf("> underruns=%-4v buflen=%-4v tickavg=%-12s drift=%-12s | byte=%-4v freq=%v\n",
			al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox(), data, freq)
		mod.SetFreq(freq, nil)
	}
}
