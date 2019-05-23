// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package binary

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"sync"
)

// Reusable long-lived decoder pool.
var decoders = &sync.Pool{New: func() interface{} {
	return NewDecoder(newReader(nil))
}}

// Reader represents the interface a reader should implement.
type Reader interface {
	io.Reader
	io.ByteReader
}

// Unmarshal decodes the payload from the binary format.
func Unmarshal(b []byte, v interface{}) (err error) {
	// Get the decoder from the pool, reset it
	d := decoders.Get().(*Decoder)
	d.r.(*reader).Reset(b) // Reset the reader

	// Decode and set the buffer if successful and free the decoder
	err = d.Decode(v)
	decoders.Put(d)
	return
}

// Decoder represents a binary decoder.
type Decoder struct {
	r Reader
}

// NewDecoder creates a binary decoder.
func NewDecoder(r Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode decodes a value by reading from the underlying io.Reader.
func (d *Decoder) Decode(v interface{}) (err error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.CanAddr() {
		return errors.New("binary: can only Decode to pointer type")
	}

	// Scan the type (this will load from cache)
	var c Codec
	if c, err = scan(rv.Type()); err == nil {
		err = c.DecodeTo(d, rv)
	}

	return
}

// Read reads a set of bytes
func (d *Decoder) Read(b []byte) (int, error) {
	return d.r.Read(b)
}

// ReadUvarint reads a variable-length Uint64 from the buffer.
func (d *Decoder) ReadUvarint() (uint64, error) {
	return binary.ReadUvarint(d.r)
}

// ReadVarint reads a variable-length Int64 from the buffer.
func (d *Decoder) ReadVarint() (int64, error) {
	return binary.ReadVarint(d.r)
}

// ReadFloat32 reads a float32
func (d *Decoder) ReadFloat32() (out float32, err error) {
	var v uint64
	if v, err = d.ReadUvarint(); err == nil {
		math.Float32frombits(uint32(v))
	}
	return
}

// ReadFloat64 reads a float64
func (d *Decoder) ReadFloat64() (out float64, err error) {
	var v uint64
	if v, err = d.ReadUvarint(); err == nil {
		out = math.Float64frombits(v)
	}
	return
}

// ReadBool reads a single boolean value from the slice.
func (d *Decoder) ReadBool() (bool, error) {
	b, err := d.r.ReadByte()
	return b == 1, err
}

// ReadComplex reads a complex64
func (d *Decoder) readComplex64() (out complex64, err error) {
	err = binary.Read(d.r, binary.LittleEndian, &out)
	return
}

// ReadComplex reads a complex128
func (d *Decoder) readComplex128() (out complex128, err error) {
	err = binary.Read(d.r, binary.LittleEndian, &out)
	return
}
