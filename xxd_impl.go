package xxd

import (
	"bufio"
	"io"
	"log"
	"strconv"
)

const (
	ldigits = "0123456789abcdef"
	udigits = "0123456789ABCDEF"
)

func XxdBasic(r io.Reader, w io.Writer, xxdCfg *Config) error {
	return Xxd(r, w, "-", xxdCfg)

}

func Xxd(r io.Reader, w io.Writer, fname string, xxdCfg *Config) error {
	var (
		lineOffset int64
		hexOffset  = make([]byte, 6)
		groupSize  int
		cols       int
		octs       int
		caps       = ldigits
		doCHeader  = true
		doCEnd     bool
		// enough room for "unsigned char NAME_FORMAT[] = {"
		varDeclChar = make([]byte, 14+len(fname)+6)
		// enough room for "unsigned int NAME_FORMAT = "
		varDeclInt = make([]byte, 16+len(fname)+7)
		nulLine    int64
		totalOcts  int
	)

	// Generate the first and last line in the -i output:
	// e.g. unsigned char foo_txt[] = { and unsigned int foo_txt_len =
	if xxdCfg.DumpType == DumpCformat {
		// copy over "unnsigned char " and "unsigned int"
		_ = copy(varDeclChar[0:14], unsignedChar[:])
		_ = copy(varDeclInt[0:16], unsignedInt[:])

		for i := 0; i < len(fname); i++ {
			if fname[i] != '.' {
				varDeclChar[14+i] = fname[i]
				varDeclInt[16+i] = fname[i]
			} else {
				varDeclChar[14+i] = '_'
				varDeclInt[16+i] = '_'
			}
		}
		// copy over "[] = {" and "_len = "
		_ = copy(varDeclChar[14+len(fname):], brackets[:])
		_ = copy(varDeclInt[16+len(fname):], lenEquals[:])
	}

	// Switch between upper- and lower-case hex chars
	if xxdCfg.Upper {
		caps = udigits
	}

	// xxd -bpi FILE outputs in binary format
	// xxd -b -p -i FILE outputs in C format
	// simply catch the last option since that's what I assume the author
	// wanted...
	if xxdCfg.Columns == -1 {
		switch dumpType {
		case DumpPostscript:
			cols = 30
		case DumpCformat:
			cols = 12
		case DumpBinary:
			cols = 6
		default:
			cols = 16
		}
	} else {
		cols = xxdCfg.Columns
	}

	// See above comment
	switch dumpType {
	case DumpBinary:
		octs = 8
		groupSize = 1
	case DumpPostscript:
		octs = 0
	case DumpCformat:
		octs = 4
	default:
		octs = 2
		groupSize = 2
	}

	if xxdCfg.Group != -1 {
		groupSize = xxdCfg.Group
	}

	// If -l is smaller than the number of cols just truncate the cols
	if xxdCfg.Length != -1 {
		if xxdCfg.Length < int(cols) {
			cols = int(xxdCfg.Length)
		}
	}

	if octs < 1 {
		octs = cols
	}

	// These are bumped down from the beginning of the function in order to
	// allow for their sizes to be allocated based on the user's specification
	var (
		line = make([]byte, cols)
		char = make([]byte, octs)
	)

	c := int64(0) // number of characters
	nl := int64(0)
	r = bufio.NewReader(r)

	var (
		v   byte
		n   int
		err error
	)

	for {
		n, err = io.ReadFull(r, line)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
		// Speed it up a bit ;)
		if dumpType == DumpPostscript && n != 0 {
			// Post script values
			// Basically just raw hex output
			for i := 0; i < n; i++ {
				hexEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++
			}
			continue
		}

		if n == 0 {
			if dumpType == DumpPostscript {
				w.Write(newLine)
			}

			if dumpType == DumpCformat {
				doCEnd = true
			} else {
				return nil // Hidden return!
			}
		}

		if xxdCfg.Length != -1 {
			if totalOcts == xxdCfg.Length {
				break
			}
			totalOcts += xxdCfg.Length
		}

		if xxdCfg.AutoSkip && empty(&line) {
			if nulLine == 1 {
				w.Write(asterisk)
				w.Write(newLine)
			}

			nulLine++

			if nulLine > 1 {
				lineOffset++ // continue to increment our offset
				continue
			}
		}

		if xxdCfg.DumpType <= DumpBinary { // either hex or binary
			// Line offset
			hexOffset = strconv.AppendInt(hexOffset[0:0], lineOffset, 16)
			w.Write(zeroHeader[0:(6 - len(hexOffset))])
			w.Write(hexOffset)
			w.Write(zeroHeader[6:])
			lineOffset++
		} else if doCHeader {
			w.Write(varDeclChar)
			w.Write(newLine)
			doCHeader = false
		}

		if xxdCfg.DumpType == DumpBinary {
			// Binary values
			for i, k := 0, octs; i < n; i, k = i+1, k+octs {
				binaryEncode(char, line[i:i+1])
				w.Write(char)
				c++

				if k == octs*groupSize {
					k = 0
					w.Write(space)
				}
			}
		} else if xxdCfg.DumpType == DumpCformat {
			// C values
			if !doCEnd {
				w.Write(doubleSpace)
			}
			for i := 0; i < n; i++ {
				cfmtEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++

				// don't add spaces to EOL
				if i != n-1 {
					w.Write(commaSpace)
				} else if n == cols {
					w.Write(comma)
				}
			}
		} else {
			// Hex values -- default xxd FILE output
			for i, k := 0, octs; i < n; i, k = i+1, k+octs {
				hexEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++

				if k == octs*groupSize {
					k = 0 // reset counter
					w.Write(space)
				}
			}
		}

		if doCEnd {
			w.Write(varDeclInt)
			w.Write([]byte(strconv.FormatInt(c, 10)))
			w.Write(semiColonNl)
			return nil
		}

		if n < len(line) && dumpType <= DumpBinary {

			//Each line should have len(line), but we have a deficit
			maxGroupsPerLine := len(line) / groupSize

			//'n' can fill up how many groups?
			k := n
			totalGroups := 1
			for k > 0 {
				k = k - groupSize

				if k >= 0 {
					totalGroups++
				}
			}
			//k will be now be the deficit of the last incomplete group

			if k < 0 {
				//incomplete group
				for ; k < 0; k++ {
					w.Write(twoSpaces)
				}
			} else if k == 0 {

				for i := 0; i < groupSize; i++ {
					w.Write(twoSpaces)
				}
			}
			w.Write(space)

			//finish off the rest of the deficit groups

			for i := totalGroups; i < maxGroupsPerLine; i++ {
				for i := 0; i < groupSize; i++ {
					w.Write(twoSpaces)
				}
				w.Write(space)
			}
		}

		if dumpType != DumpCformat {
			w.Write(space)
		}

		if dumpType <= DumpBinary {

			w.Write(space)

			// Character values

			b := line[:n]

			// |hello, world!| instead of hello, world!
			if xxdCfg.Bars {
				w.Write(bar)
			}
			// EBCDIC
			if xxdCfg.Ebcdic {
				for i := 0; i < len(b); i++ {
					v = b[i]
					if v >= ebcdicOffset {
						e := ebcdicTable[v-ebcdicOffset : v-ebcdicOffset+1]
						if e[0] > 0x1f && e[0] < 0x7f {
							w.Write(e)
						} else {
							w.Write(dot)
						}
					} else {
						w.Write(dot)
					}
				}
				if xxdCfg.Bars {
					w.Write(bar)
				}
				// ASCII
			} else {
				var v byte
				for i := 0; i < len(b); i++ {
					v = b[i]
					if v > 0x1f && v < 0x7f {
						w.Write(line[i : i+1])
					} else {
						w.Write(dot)
					}
				}
			}

			if xxdCfg.Bars {
				w.Write(bar)
			}

		}
		w.Write(newLine)
		nl++
	}
	return nil
}

// convert a byte into its binary representation
func binaryEncode(dst, src []byte) {
	d := uint(0)
	_, _ = src[0], dst[7]
	for i := 7; i >= 0; i-- {
		if src[0]&(1<<d) == 0 {
			dst[i] = '0'
		} else {
			dst[i] = '1'
		}
		d++
	}
}

// returns -1 on success
// returns k > -1 if space found where k is index of space byte
func binaryDecode(dst, src []byte) int {
	var v, d byte

	for i := 0; i < len(src); i++ {
		v, d = src[i], d<<1
		if isSpace(v) { // found a space, so between groups
			if i == 0 {
				return 1
			}
			return i
		}
		if v == '1' {
			d ^= 1
		} else if v != '0' {
			return i // will catch issues like "000000: "
		}
	}

	dst[0] = d
	return -1
}

func cfmtEncode(dst, src []byte, hextable string) {
	b := src[0]
	dst[3] = hextable[b&0x0f]
	dst[2] = hextable[b>>4]
	dst[1] = 'x'
	dst[0] = '0'
}

// copied from encoding/hex package in order to add support for uppercase hex
func hexEncode(dst, src []byte, hextable string) {
	b := src[0]
	dst[1] = hextable[b&0x0f]
	dst[0] = hextable[b>>4]
}

// copied from encoding/hex package
// returns -1 on bad byte or space (\t \s \n)
// returns -2 on two consecutive spaces
// returns 0 on success
func hexDecode(dst, src []byte) int {
	_, _ = src[2], dst[0]

	if isSpace(src[0]) {
		if isSpace(src[1]) {
			return -2
		}
		return -1
	}

	if isPrefix(src[0:2]) {
		src = src[2:]
	}

	for i := 0; i < len(src)/2; i++ {
		a, ok := fromHexChar(src[i*2])
		if !ok {
			return -1
		}
		b, ok := fromHexChar(src[i*2+1])
		if !ok {
			return -1
		}

		dst[0] = (a << 4) | b
	}
	return 0
}

// copied from encoding/hex package
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}

// check if entire line is full of empty []byte{0} bytes (nul in C)
func empty(b *[]byte) bool {
	for i := 0; i < len(*b); i++ {
		if (*b)[i] != 0 {
			return false
		}
	}
	return true
}

// quick binary tree check
// probably horribly written idk it's late at night
func parseSpecifier(b string) float64 {
	lb := len(b)
	if lb == 0 {
		return 0
	}

	var b0, b1 byte
	if lb < 2 {
		b0 = b[0]
		b1 = '0'
	} else {
		b1 = b[1]
		b0 = b[0]
	}

	if b1 != '0' {
		if b1 == 'b' { // bits, so convert bytes to bits for os.Seek()
			if b0 == 'k' || b0 == 'K' {
				return 0.0078125
			}

			if b0 == 'm' || b0 == 'M' {
				return 7.62939453125e-06
			}

			if b0 == 'g' || b0 == 'G' {
				return 7.45058059692383e-09
			}
		}

		if b1 == 'B' { // kilo/mega/giga- bytes are assumed
			if b0 == 'k' || b0 == 'K' {
				return 1024
			}

			if b0 == 'm' || b0 == 'M' {
				return 1048576
			}

			if b0 == 'g' || b0 == 'G' {
				return 1073741824
			}
		}
	} else { // kilo/mega/giga- bytes are assumed for single b, k, m, g
		if b0 == 'k' || b0 == 'K' {
			return 1024
		}

		if b0 == 'm' || b0 == 'M' {
			return 1048576
		}

		if b0 == 'g' || b0 == 'G' {
			return 1073741824
		}
	}

	return 1 // assumes bytes as fallback
}

// parses *seek input
func parseSeek(s string) int64 {
	var (
		sl    = len(s)
		split int
	)

	switch {
	case sl >= 2:
		if sl == 2 {
			split = 1
		} else {
			split = 2
		}
	case sl != 0:
		split = 0
	default:
		log.Fatalln("seek string somehow has len of 0")
	}

	mod := parseSpecifier(s[sl-split:])
	ret, err := strconv.ParseFloat(s[:sl-split], 64) // 64 bit float
	if err != nil {
		log.Fatalln(err)
	}

	return int64(ret * mod)
}

// is byte a space? (\t, \n, \s)
func isSpace(b byte) bool {
	switch b {
	case 32, 12, 9:
		return true
	default:
		return false
	}
}

// are the two bytes hex prefixes? (0x or 0X)
func isPrefix(b []byte) bool {
	return b[0] == '0' && (b[1] == 'x' || b[1] == 'X')
}
