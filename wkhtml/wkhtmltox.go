package wkhtml

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type ToX struct {
}

const wkhtmltopdf = "wkhtmltopdf"

var wkCh1, wkChN chan *InOut

func init() {
	options := ExecOptions{Timeout: 10 * time.Second}
	wkCh1 = make(chan *InOut, 1)
	wkChN = make(chan *InOut, runtime.NumCPU()*2)
	go func() {
		for {
			wk, err := options.NewPrepare(wkhtmltopdf, "--read-args-from-stdin")
			if err != nil {
				time.Sleep(10 * time.Second)
				continue
			}
			wkCh1 <- wk
		}
	}()
}

func getWk() *InOut {
	select {
	case wk := <-wkChN:
		return wk
	default:
		return <-wkCh1
	}
}

func backToPool(wk *InOut) {
	select {
	case wkChN <- wk:
		return
	default:
		if err := wk.Kill(); err != nil {
			log.Printf("failed to kill, error: %v", err)
		}
		return
	}
}

func (p *ToX) ToPDFStdinArgs(htmlURL, extraArgs string) (pdf []byte, err error) {
	wk := getWk()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	out := filepath.Join(dir, NewUUID().String())
	in := htmlURL + " " + out + "\n"
	if extraArgs != "" {
		in = extraArgs + " " + in
	}
	result, err := wk.Send(in, "Done", "Error:")
	log.Printf("wk result: %s", result)
	if err == nil {
		pdf, err = os.ReadFile(out)
		os.Remove(out)
		return pdf, err
	}

	if err == ErrTimeout {
		if err := wk.Kill(); err != nil {
			log.Printf("failed to kill, error: %v", err)
		}
	} else {
		backToPool(wk)
	}
	return nil, err
}

func (p *ToX) ToPDFByURL(htmlURL, extraArgs string) (pdf []byte, err error) {
	cmd := wkhtmltopdf + extraArgs + " --quiet " + htmlURL + " -"
	options := ExecOptions{Timeout: 10 * time.Second}
	return options.Exec(nil, "sh", "-c", cmd)
}

func (p *ToX) ToPDF(html []byte) (pdf []byte, err error) {
	args := []string{"--quiet", "-", "-"}
	options := ExecOptions{Timeout: 10 * time.Second}
	return options.Exec(html, wkhtmltopdf, args...)
}

// NewUUID creates a new random UUID or panics.
func NewUUID() UUID {
	return MustNewUUID(NewRandomUUID())
}

// MustNewUUID returns uuid if err is nil and panics otherwise.
func MustNewUUID(uuid UUID, err error) UUID {
	if err != nil {
		panic(err)
	}
	return uuid
}

// String returns the string form of uuid, xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// , or "" if uuid is invalid.
func (uuid UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return string(buf[:])
}

func encodeHex(dst []byte, uuid UUID) {
	hex.Encode(dst, uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}

var rander = rand.Reader // random function
var Nil UUID             // empty UUID, all zeros

// A UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC 4122.
type UUID [16]byte

// NewRandomUUID returns a Random (Version 4) UUID.
//
// The strength of the UUIDs is based on the strength of the crypto/rand
// package.
//
// A note about uniqueness derived from the UUID Wikipedia entry:
//
//  Randomly generated UUIDs have 122 random bits.  One's annual risk of being
//  hit by a meteorite is estimated to be one chance in 17 billion, that
//  means the probability is about 0.00000000006 (6 × 10−11),
//  equivalent to the odds of creating a few tens of trillions of UUIDs in a
//  year and having one duplicate.
func NewRandomUUID() (UUID, error) {
	return NewRandomUUIDFromReader(rander)
}

// NewRandomUUIDFromReader returns a UUID based on bytes read from a given io.Reader.
func NewRandomUUIDFromReader(r io.Reader) (UUID, error) {
	var uuid UUID
	_, err := io.ReadFull(r, uuid[:])
	if err != nil {
		return Nil, err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid, nil
}
