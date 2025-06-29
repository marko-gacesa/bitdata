// Copyright (c) 2025 by Marko Gaćeša

package bitdata

import (
	"errors"
	"io"
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

func (w *Writer) BitData() BitData {
	return *w.data
}

func (w *Writer) WriteBool(v bool) {
	if v {
		write[byte](w, 1, 1)
	} else {
		write[byte](w, 0, 1)
	}
}

func (w *Writer) Write8(v uint8, bitCount byte) {
	write[uint8](w, v, bitCount)
}

func (w *Writer) Write16(v uint16, bitCount byte) {
	write[uint16](w, v, bitCount)
}

func (w *Writer) Write32(v uint32, bitCount byte) {
	write[uint32](w, v, bitCount)
}

func (w *Writer) Write64(v uint64, bitCount byte) {
	write[uint64](w, v, bitCount)
}

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

func (r *Reader) Skip(bitCount uint) {
	r.bitsRead += bitCount
}

func (r *Reader) ReadBool() (bool, error) {
	v, err := read[byte](r, 1)
	if err != nil {
		return false, err
	}

	return v != 0, nil
}

func (r *Reader) Read8(bitCount byte) (uint8, error) {
	if bitCount > 8 {
		return 0, ErrBitCountTooBig
	}
	return read[uint8](r, bitCount)
}

func (r *Reader) Read16(bitCount byte) (uint16, error) {
	if bitCount > 16 {
		return 0, ErrBitCountTooBig
	}
	return read[uint16](r, bitCount)
}

func (r *Reader) Read32(bitCount byte) (uint32, error) {
	if bitCount > 32 {
		return 0, ErrBitCountTooBig
	}
	return read[uint32](r, bitCount)
}

func (r *Reader) Read64(bitCount byte) (uint64, error) {
	if bitCount > 64 {
		return 0, ErrBitCountTooBig
	}
	return read[uint64](r, bitCount)
}

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

func (r *ReaderError) ReadBool() (v bool) {
	if r.err == nil {
		v, r.err = r.reader.ReadBool()
	}
	return
}

func (r *ReaderError) Read8(bitCount byte) (v uint8) {
	if r.err == nil {
		v, r.err = r.reader.Read8(bitCount)
	}
	return
}

func (r *ReaderError) Read16(bitCount byte) (v uint16) {
	if r.err == nil {
		v, r.err = r.reader.Read16(bitCount)
	}
	return
}

func (r *ReaderError) Read32(bitCount byte) (v uint32) {
	if r.err == nil {
		v, r.err = r.reader.Read32(bitCount)
	}
	return
}

func (r *ReaderError) Read64(bitCount byte) (v uint64) {
	if r.err == nil {
		v, r.err = r.reader.Read64(bitCount)
	}
	return
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func mask[T integer](n byte) T {
	return T(1)<<n - 1
}

func write[T integer](w *Writer, value T, bitCount byte) {
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

func read[T integer](r *Reader, bitCount byte) (T, error) {
	if bitCount == 0 {
		return 0, nil
	}

	var value T

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
