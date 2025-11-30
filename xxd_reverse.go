package xxd

import (
	"bufio"
	"io"
)

func XxdReverse(r io.Reader, w io.Writer, xxdCfg *Config) error {
	var (
		cols int
		octs int
		char = make([]byte, 1)
	)

	if xxdCfg.Columns != -1 {
		cols = xxdCfg.Columns
	}

	switch dumpType {
	case DumpBinary:
		octs = 8
	case DumpCformat:
		octs = 4
	default:
		octs = 2
	}

	if xxdCfg.Length != -1 {
		if xxdCfg.Length < cols {
			cols = xxdCfg.Length
		}
	}

	if octs < 1 {
		octs = cols
	}

	c := int64(0) // number of characters
	rd := bufio.NewReader(r)
	for {
		line, err := rd.ReadBytes('\n') // read up until a newline
		n := len(line)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}

		if n == 0 {
			return nil
		}

		if dumpType == DumpHex {
			for i := 0; n >= octs; {
				if rv := hexDecode(char, line[i:i+octs]); rv == 0 {
					w.Write(char)
					i += 2
					n -= 2
					c++
				} else if rv == -1 {
					i++
					n--
				} else { // if rv == -2
					i += 2
					n -= 2
				}
			}
		} else if dumpType == DumpBinary {
			for i := 0; n >= octs; {
				if binaryDecode(char, line[i:i+octs]) != -1 {
					i++
					n--
					continue
				} else {
					w.Write(char)
					i += 8
					n -= 8
					c++
				}
			}
		} else if dumpType == DumpPostscript {
			for i := 0; n >= octs; i++ {
				if hexDecode(char, line[i:i+octs]) == 0 {
					w.Write(char)
					c++
				}
				n--
			}
		} else if dumpType == DumpCformat {
			for i := 0; n >= octs; {
				if rv := hexDecode(char, line[i:i+octs]); rv == 0 {
					w.Write(char)
					i += 4
					n -= 4
					c++
				} else if rv == -1 {
					i++
					n--
				} else { // if rv == -2
					i += 2
					n -= 2
				}
			}
		}

		// For some reason "xxd FILE | xxd -r -c N" truncates the output,
		// so we'll do it as well
		// "xxd FILE | xxd -r -l N" doesn't truncate
		if c == int64(cols) && cols > 0 {
			return nil
		}
	}
}
