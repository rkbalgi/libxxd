package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	flag "github.com/ogier/pflag"
	xxd "github.com/rkbalgi/libxxd"
)

// cli flags

// usage and version
const (
	Help = `Usage:
       xxd [options] [infile [outfile]]
    or
       xxd -r [-s offset] [-c cols] [--ps] [infile [outfile]]
Options:
    -a, --autoskip     toggle autoskip: A single '*' replaces nul-lines. Default off.
    -B, --bars         print pipes/bars before/after ASCII/EBCDIC output. Default off.
    -b, --binary       binary digit dump (incompatible with -ps, -i, -r). Default hex.
    -c, --cols         format <cols> octets per line. Default 16 (-i 12, --ps 30).
    -E, --ebcdic       show characters in EBCDIC. Default ASCII.
    -g, --groups       number of octets per group in normal output. Default 2.
    -h, --help         print this summary.
    -i, --include      output in C include file style.
    -l, --length       stop after <len> octets.
    -p, --ps           output in postscript plain hexdump style.
    -r, --reverse      reverse operation: convert (or patch) hexdump into ASCII output.
                       * reversing non-hexdump formats require -r<flag> (i.e. -rb, -ri, -rp).
    -s, --seek         start at <seek> bytes/bits in file. Byte/bit postfixes can be used.
    		       * byte/bit postfix units are multiples of 1024.
    		       * bits (kb, mb, etc.) will be rounded down to nearest byte.
    -u, --uppercase    use upper case hex letters.
    -v, --version      show version.`
	Version = `xxd v2.0 2014-17-01 by Felix Geisendörfer and Eric Lagergren`
)

var (
	autoskip   = flag.BoolP("autoskip", "a", false, "toggle autoskip (* replaces nul lines")
	bars       = flag.BoolP("bars", "B", false, "print |ascii| instead of ascii")
	binary     = flag.BoolP("binary", "b", false, "binary dump, incompatible with -ps, -i, -r")
	columns    = flag.IntP("cols", "c", -1, "format <cols> octets per line")
	ebcdic     = flag.BoolP("ebcdic", "E", false, "use EBCDIC instead of ASCII")
	group      = flag.IntP("group", "g", -1, "num of octets per group")
	cfmt       = flag.BoolP("include", "i", false, "output in C include format")
	length     = flag.Int64P("len", "l", -1, "stop after len octets")
	postscript = flag.BoolP("ps", "p", false, "output in postscript plain hd style")
	reverse    = flag.BoolP("reverse", "r", false, "convert hex to binary")
	seek       = flag.StringP("seek", "s", "", "start at seek bytes abs")
	upper      = flag.BoolP("uppercase", "u", false, "use uppercase hex letters")
	version    = flag.BoolP("version", "v", false, "print version")
)

func main() {
	xxdCfg := XxdConfig{}

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, Help)
		os.Exit(0)
	}
	flag.Parse()

	if *version {
		fmt.Fprintln(os.Stderr, Version)
		os.Exit(0)
	}

	if flag.NArg() > 2 {
		log.Fatalf("Too many arguments after %s\n", flag.Args()[1])
	}

	var (
		err  error
		file string
	)

	if flag.NArg() >= 1 {
		file = flag.Args()[0]
	} else {
		file = "-"
	}

	var inFile *os.File
	if file == "-" {
		inFile = os.Stdin
		file = "stdin"
	} else {
		inFile, err = os.Open(file)
		if err != nil {
			log.Fatalln(err)
		}
	}
	defer inFile.Close()

	// Start *seek bytes into file
	if *seek != "" {
		sv := parseSeek(*seek)
		_, err = inFile.Seek(sv, os.SEEK_SET)
		if err != nil {
			log.Fatalln(err)
		}
	}

	var outFile *os.File
	if flag.NArg() == 2 {
		outFile, err = os.Create(flag.Args()[1])
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		outFile = os.Stdout
	}
	defer outFile.Close()

	switch {
	case *binary:
		xxdCfg.DumpType = xxd.DumpBinary
	case *cfmt:
		xxdCfg.dumpType = xxd.DumpCformat
	case *postscript:
		xxdCfg.dumpType = xxd.DumpPostscript
	default:
		xxdCfg.dumpType = xxd.DumpHex
	}

	out := bufio.NewWriter(outFile)
	defer out.Flush()

	if *reverse {
		if err = xxd.XxdReverse(inFile, out, xxdCfg); err != nil {
			log.Fatalln(err)
		}
		return
	}

	if err = xxd.Xxd(inFile, out, file, xxdCfg); err != nil {
		log.Fatalln(err)
	}
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
