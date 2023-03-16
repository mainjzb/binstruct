package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/ghostiam/binstruct"
)

// Portable Network Graphics (PNG) Specification: https://www.w3.org/TR/PNG/
func main() {
	file, err := os.Open("sample.png")
	if err != nil {
		log.Fatal(err)
	}

	var png PNG
	decoder := binstruct.NewDecoder(file, binary.BigEndian)
	// decoder.SetDebug(true)
	err = decoder.Decode(&png)
	if err != nil {
		log.Fatal(err)
	}

	spew.Dump(png)
}

type PNG struct {
	Header [8]byte // Magic: 0x89 0x50 0x4E 0x47 0x0D 0x0A 0x1A 0x0A
	IHDR   IHDR
	Chunks []Chunk `bin:"ReadChunks"`
}

func (png *PNG) ReadChunks(r binstruct.Reader) error {
	for {
		var c Chunk
		err := r.Unmarshal(&c)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed read png chunk: %w", err)
		}

		png.Chunks = append(png.Chunks, c)
	}
}

// https://www.w3.org/TR/PNG/#11IHDR
type IHDR struct {
	Len  int32
	Type string `bin:"len:4"`
	Data IHDRData
	CRC  [4]byte
}

type IHDRData struct {
	Width             int32
	Height            int32
	BitDepth          int8
	ColorType         PNGColorType
	CompressionMethod int8
	FilterMethod      int8
	InterlaceMethod   int8
}

type PNGColorType int8

func (t PNGColorType) String() string {
	colors := [...]string{
		"Greyscale",            // 0
		"Invalid",              // 1
		"Truecolor",            // 2
		"Indexed",              // 3
		"Greyscale with alpha", // 4
		"Invalid",              // 5
		"Truecolor with alpha", // 6
	}
	return colors[t]
}

type Chunk struct {
	Len  int32
	Type string      `bin:"len:4"`
	Data interface{} `bin:"ReadChunkData"`
	CRC  [4]byte
}

func (c *Chunk) ReadChunkData(r binstruct.Reader) (interface{}, error) {
	switch c.Type {
	case "PLTE": // https://www.w3.org/TR/PNG/#11PLTE
		v := PaletteData{DataLen: c.Len}
		err := r.Unmarshal(&v)
		return v, err
	case "cHRM": // https://www.w3.org/TR/PNG/#11cHRM
		var v ChromaticData
		err := r.Unmarshal(&v)
		return v, err
	case "gAMA": // https://www.w3.org/TR/PNG/#11gAMA
		gamma, err := r.ReadInt32()
		return GammaData(gamma), err
		// case "sRGB": // https://www.w3.org/TR/PNG/#11sRGB
		// case "bKGD": // https://www.w3.org/TR/PNG/#11bKGD
	case "pHYs": // https://www.w3.org/TR/PNG/#11pHYs
		var v PhysicalPixelData
		err := r.Unmarshal(&v)
		return v, err

		// case "tIME": // https://www.w3.org/TR/PNG/#11tIME
	case "iTXt": // https://www.w3.org/TR/PNG/#11iTXt
		v := InternationalTextData{DataLen: c.Len}
		err := r.Unmarshal(&v)
		return v, err
		// case "tEXt": // https://www.w3.org/TR/PNG/#11tEXt
		// case "zTXt": // https://www.w3.org/TR/PNG/#11zTXt
	default:
		// IDAT, IEND and others
		// read raw
		_, b, err := r.ReadBytes(int(c.Len))
		return RawData(b), err
	}
}

type RawData []byte

func (RawData) String() string {
	return "<skip>"
}

type PaletteData struct {
	DataLen int32 `bin:"-"` // Helper
	Entries []RGB `bin:"len:DataLen/3"`
}

type RGB struct {
	R int8
	G int8
	B int8
}

type ChromaticData struct {
	White Point
	Red   Point
	Green Point
	Blue  Point
}

type Point struct {
	X int32
	Y int32
}

type GammaData int32

type PhysicalPixelData struct {
	Point
	Unit int8
}

type InternationalTextData struct {
	DataLen           int32  `bin:"-"` // Helper
	Keyword           string `bin:"NullTerminatedString"`
	CompressionFlag   byte
	CompressionMethod byte
	LanguageTag       string `bin:"NullTerminatedString"`
	TranslatedKeyword string `bin:"NullTerminatedString"`
	Text              string `bin:"len:DataLen-2"` // DataLen - CompressionFlag - CompressionMethod
}

func (d *InternationalTextData) NullTerminatedString(r binstruct.Reader) (string, error) {
	var b []byte

	var readCount int32
	for {
		readByte, err := r.ReadByte()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}

		readCount++

		if readByte == 0x00 {
			break
		}

		b = append(b, readByte)
	}

	d.DataLen -= readCount

	return string(b), nil
}
