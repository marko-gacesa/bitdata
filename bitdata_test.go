// Copyright (c) 2025 by Marko Gaćeša

package bitdata

import (
	"io"
	"math/rand/v2"
	"testing"
)

func TestBitData(t *testing.T) {
	type value struct {
		data uint64
		size byte
	}
	tests := []struct {
		name    string
		data    []value
		expSize int
	}{
		{
			name: "single-byte",
			data: []value{
				{data: uint64(0b101), size: 3},
				{data: uint64(0b1001), size: 4},
				{data: uint64(1), size: 1},
			},
			expSize: 1,
		},
		{
			name: "two-bytes",
			data: []value{
				{data: uint64(0b11011), size: 5},
				{data: uint64(0b10001), size: 5},
				{data: uint64(0b110011), size: 6},
			},
			expSize: 2,
		},
		{
			name: "big-values",
			data: []value{
				{data: uint64(0xDEADBEEFDEAFFEED), size: 64},
				{data: uint64(0xCEED), size: 16},
				{data: uint64(0xFEEDDEAD), size: 32},
			},
			expSize: 8 + 2 + 4,
		},
		{
			name: "big-values-offset",
			data: []value{
				{data: uint64(0b10010), size: 5},
				{data: uint64(0xDEADBEEFDEAFFEED), size: 64},
				{data: uint64(0xCEED), size: 16},
				{data: uint64(0xFEEDDEAD), size: 32},
			},
			expSize: 1 + 8 + 2 + 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := NewWriter()
			for _, v := range test.data {
				w.Write64(v.data, v.size)
			}

			d := w.BitData()
			if len(d) != test.expSize {
				t.Errorf("data size mismatch: want=%d got=%d", test.expSize, len(d))
			}

			r := NewReader(d)
			for i, v := range test.data {
				vv, err := r.Read64(v.size)
				if err != nil {
					t.Errorf("failed for data index %d: %s", i, err)
					return
				}
				if vv != v.data {
					t.Errorf("data mismatch for data index %d: want=%b got=%b", i, v.data, vv)
				}
			}
		})
	}
}

func TestBitDataFuzzy(t *testing.T) {
	type value struct {
		data uint64
		size byte
		bits byte
	}

	values := make([]value, 1000)

	for i := 0; i < len(values); i++ {
		bitCount := rand.N[byte](64) + 1
		var bits byte
		switch {
		case bitCount > 32:
			bits = 64
		case bitCount > 16:
			bits = 32
		case bitCount > 8:
			bits = 16
		case bitCount > 1:
			bits = 8
		default:
			bits = 1
		}
		values[i] = value{
			data: rand.Uint64() & mask[uint64](bitCount),
			size: bitCount,
			bits: bits,
		}
	}

	w := NewWriter()
	for _, v := range values {
		switch v.bits {
		case 64:
			w.Write64(v.data, v.size)
		case 32:
			w.Write32(uint32(v.data), v.size)
		case 16:
			w.Write16(uint16(v.data), v.size)
		case 8:
			w.Write8(uint8(v.data), v.size)
		case 1:
			w.WriteBool(v.data&1 == 1)
		default:
			t.Errorf("unexpected bit count: %d", v.bits)
		}
	}

	d := w.BitData()

	r := NewReader(d)
	for i, v := range values {
		var (
			v64 uint64
			v32 uint32
			v16 uint16
			v8  uint8
			v1  bool
			err error
		)
		switch v.bits {
		case 64:
			v64, err = r.Read64(v.size)
		case 32:
			v32, err = r.Read32(v.size)
			v64 = uint64(v32)
		case 16:
			v16, err = r.Read16(v.size)
			v64 = uint64(v16)
		case 8:
			v8, err = r.Read8(v.size)
			v64 = uint64(v8)
		case 1:
			v1, err = r.ReadBool()
			if v1 {
				v64++
			}
		default:
			t.Errorf("unexpected bit count: %d", v.bits)
		}
		if err != nil {
			t.Errorf("failed for data index %d: %s", i, err)
			return
		}
		if v64 != v.data {
			t.Errorf("data mismatch for data index %d: want=%b got=%b", i, v.data, v64)
		}
	}
}

func TestBitDataZero(t *testing.T) {
	w := NewWriter()
	w.Write8(0, 0)
	if w.bitsWritten != 0 {
		t.Errorf("expected 0 bits, got %d", w.bitsWritten)
	}
	if len(*w.data) != 0 {
		t.Errorf("expected 0 len, got %d", len(*w.data))
	}

	r := NewReader(w.BitData())
	_, err := r.Read8(0)
	if r.bitsRead != 0 {
		t.Errorf("expected 0 bits, got %d", r.bitsRead)
	}
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}

func TestBitDataError(t *testing.T) {
	var (
		w   *Writer
		r   *Reader
		err error
	)

	w = NewWriter()

	r = NewReader(w.BitData())
	_, err = r.ReadBool()
	if err != io.ErrUnexpectedEOF {
		t.Errorf("expected error, got %v", err)
	}

	w.Write8(0b111111, 6)
	w.Write8(0b11111, 5)
	w.Write8(0b1111, 4)

	r = NewReader(w.BitData())

	_, err = r.Read8(10)
	if err != ErrBitCountTooBig {
		t.Errorf("expected error, got %v", err)
	}

	_, err = r.Read16(20)
	if err != ErrBitCountTooBig {
		t.Errorf("expected error, got %v", err)
	}

	_, err = r.Read32(40)
	if err != ErrBitCountTooBig {
		t.Errorf("expected error, got %v", err)
	}

	_, err = r.Read64(65)
	if err != ErrBitCountTooBig {
		t.Errorf("expected error, got %v", err)
	}

	_, err = r.Read16(10)
	if err != nil {
		t.Error("unexpected error")
	}

	_, err = r.Read8(7)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("expected error, got %v", err)
	}
}

func TestBitDataTooManyBits(t *testing.T) {
	w := NewWriter()

	const value4bits = 0b1111

	// We write 10 bits - that is 2 bytes, so it should be possible to read 16 bits.
	w.Write8(value4bits, 10) // 10 bits, but writing a byte

	r := NewReader(w.BitData())

	if v, err := r.Read8(4); v != value4bits || err != nil {
		t.Errorf("value mismatch, want %b, got %b, error %v", value4bits, v, err)
	}

	if v, err := r.Read16(12); v != 0 || err != nil {
		t.Errorf("value mismatch, want %b, got %b, error %v", 0, v, err)
	}

	if _, err := r.ReadBool(); err != io.ErrUnexpectedEOF {
		t.Errorf("err mismatch, want %v, got %v", io.ErrUnexpectedEOF, err)
	}

	w.Write64(value4bits, 128-10) // 118 bits, but writing a uint64

	if a := w.BitData(); len(a) != 128/8 {
		t.Errorf("len mismatch, want %v, got %v", 128/8, len(a))
	}
}

func TestReaderError(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T, r *ReaderError)
	}{
		{
			name: "r1",
			fn:   func(t *testing.T, r *ReaderError) { r.ReadBool() },
		},
		{
			name: "r8",
			fn:   func(t *testing.T, r *ReaderError) { r.Read8(8) },
		},
		{
			name: "r16",
			fn:   func(t *testing.T, r *ReaderError) { r.Read16(8) },
		},
		{
			name: "r32",
			fn:   func(t *testing.T, r *ReaderError) { r.Read32(8) },
		},
		{
			name: "r64",
			fn:   func(t *testing.T, r *ReaderError) { r.Read64(8) },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// no error

			w := NewWriter()
			w.Write8(0xFF, 8)

			r := NewReaderError(w.BitData())
			test.fn(t, r)

			if want, got := error(nil), r.Error(); want != got {
				t.Errorf("want %v, got %v", want, got)
			}

			// with error

			w = NewWriter()

			r = NewReaderError(w.BitData())
			r.Skip(12)
			test.fn(t, r)

			if want, got := io.ErrUnexpectedEOF, r.Error(); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		})
	}
}
