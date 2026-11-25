// Copyright (c) 2025, 2026 by Marko Gaćeša

package bitdata

import (
	"errors"
	"io"
	"unsafe"
)

type BitData []byte

type Writer struct {
	data        *BitData
	bitsWritten uint
}

var ErrBitCountTooBig = errors.New("bit count too big")

func NewWriter() *Writer {
	return &Writer{
		data:        (*BitData)(new([]byte)),
		bitsWritten: 0,
	}
}

func (w *Writer) BitData() BitData { return *w.data }

func (w *Writer) Write[T integer](value T, bitCount byte) {
	if bitCount == 0 {
		return
	}

	value = value & mask[T](bitCount)

	idx := w.bitsWritten / 8
	ofs := w.bitsWritten % 8
	bitsRemain := int8(bitCount)

	if ofs > 0 {
		(*w.data)[idx] = (*w.data)[idx] | byte(value<<ofs)
		bits := int8(8 - byte(ofs))
		bitsRemain -= bits
		value >>= bits
	}

	for bitsRemain > 0 {
		*w.data = append(*w.data, byte(value))
		value >>= 8
		bitsRemain -= 8
	}

	w.bitsWritten += uint(bitCount)
}

func (w *Writer) WriteBool(v bool) {
	if v {
		w.Write[byte](1, 1)
	} else {
		w.Write[byte](0, 1)
	}
}

func (w *Writer) Write8(v uint8, bitCount byte)   { w.Write[uint8](v, bitCount) }
func (w *Writer) Write16(v uint16, bitCount byte) { w.Write[uint16](v, bitCount) }
func (w *Writer) Write32(v uint32, bitCount byte) { w.Write[uint32](v, bitCount) }
func (w *Writer) Write64(v uint64, bitCount byte) { w.Write[uint64](v, bitCount) }

type Reader struct {
	data     BitData
	bitsRead uint
}

func NewReader(data BitData) *Reader {
	return &Reader{
		data:     data,
		bitsRead: 0,
	}
}

func (r *Reader) Skip(bitCount uint) { r.bitsRead += bitCount }

func (r *Reader) Read[T integer](bitCount byte) (T, error) {
	if bitCount == 0 {
		return 0, nil
	}

	var value T

	if uintptr(bitCount) > unsafe.Sizeof(value)*8 {
		return 0, ErrBitCountTooBig
	}

	idx := r.bitsRead / 8
	ofs := r.bitsRead % 8
	bitsRemain := int8(bitCount)
	var bitsRead byte

	if ofs > 0 {
		if idx >= uint(len(r.data)) {
			return 0, io.ErrUnexpectedEOF
		}

		value = T(r.data[idx]>>ofs) & mask[T](bitCount)
		bits := int8(8 - byte(ofs))
		bitsRemain -= bits
		bitsRead += byte(bits)
		idx++
	}

	for bitsRemain > 0 {
		if idx >= uint(len(r.data)) {
			return 0, io.ErrUnexpectedEOF
		}

		v := T(r.data[idx]) << bitsRead
		m := mask[T](byte(bitsRemain)) << bitsRead
		value |= v & m

		idx++
		bitsRead += 8
		bitsRemain -= 8
	}

	r.bitsRead += uint(bitCount)

	return value, nil
}

func (r *Reader) ReadBool() (bool, error) {
	v, err := r.Read[byte](1)
	if err != nil {
		return false, err
	}

	return v != 0, nil
}

func (r *Reader) Read8(bitCount byte) (uint8, error)   { return r.Read[uint8](bitCount) }
func (r *Reader) Read16(bitCount byte) (uint16, error) { return r.Read[uint16](bitCount) }
func (r *Reader) Read32(bitCount byte) (uint32, error) { return r.Read[uint32](bitCount) }
func (r *Reader) Read64(bitCount byte) (uint64, error) { return r.Read[uint64](bitCount) }

type ReaderError struct {
	reader Reader
	err    error
}

func NewReaderError(data BitData) *ReaderError {
	return &ReaderError{
		reader: *NewReader(data),
		err:    nil,
	}
}

func (r *ReaderError) Error() error {
	return r.err
}

func (r *ReaderError) Skip(bitCount uint) {
	r.reader.Skip(bitCount)
}

func (r *ReaderError) Read[T integer](bitCount byte) (v T) {
	if r.err == nil {
		v, r.err = r.reader.Read[T](bitCount)
	}
	return
}

func (r *ReaderError) ReadBool() (v bool) {
	if r.err == nil {
		v, r.err = r.reader.ReadBool()
	}
	return
}

func (r *ReaderError) Read8(bitCount byte) (v uint8) {
	if r.err == nil {
		v, r.err = r.reader.Read[byte](bitCount)
	}
	return
}

func (r *ReaderError) Read16(bitCount byte) (v uint16) {
	if r.err == nil {
		v, r.err = r.reader.Read[uint16](bitCount)
	}
	return
}

func (r *ReaderError) Read32(bitCount byte) (v uint32) {
	if r.err == nil {
		v, r.err = r.reader.Read[uint32](bitCount)
	}
	return
}

func (r *ReaderError) Read64(bitCount byte) (v uint64) {
	if r.err == nil {
		v, r.err = r.reader.Read[uint64](bitCount)
	}
	return
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func mask[T integer](n byte) T {
	return T(1)<<n - 1
}
