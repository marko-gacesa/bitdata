# BitData

BitData is a simple library for packing and unpacking bit-level data to and from byte slices.

Writing data:

```go
	day := 29
	month := 6
	year := 2025

	w := NewWriter()
	w.Write8(byte(day), 5)      // 5 bits for day (1..31)
	w.Write8(byte(month), 4)    // 4 bits for month (1..12)
	w.Write16(uint16(year), 12) // 12 bits for year (1..4095)

	packedDate := w.BitData() // packedDate holds 3 bytes, or 21 used and 3 unused bits

```

Reading data:

There are two readers: `Reader` and `ReaderError`. The `Reader`'s methods return an error with each call, whereas the `ReaderError`'s methods do not return errors individually — you must check for errors at the end instead.

```go
	r := NewReaderError(packedDate)
	day = int(r.Read16(5))   // To read data use ReadBool, Read8, Read16,
	month = int(r.Read8(4))  // Read32 or Read64. The only difference is
	year = int(r.Read32(12)) // the number of bits that we can read at once.
	if err := r.Error(); err != nil {
		return errors.New("failed to unpack date")
	}

	fmt.Println(day, month, year) // Prints: 29 6 2025
```

