package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	flag "github.com/ogier/pflag"
)

// cli flags
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

func main() {
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
		dumpType = dumpBinary
	case *cfmt:
		dumpType = dumpCformat
	case *postscript:
		dumpType = dumpPostscript
	default:
		dumpType = dumpHex
	}

	out := bufio.NewWriter(outFile)
	defer out.Flush()

	if *reverse {
		if err = xxdReverse(inFile, out); err != nil {
			log.Fatalln(err)
		}
		return
	}

	if err = xxd(inFile, out, file); err != nil {
		log.Fatalln(err)
	}
}
