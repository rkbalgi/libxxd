package xxd

const (
	DumpHex = iota
	DumpBinary
	DumpCformat
	DumpPostscript
)

const ebcdicOffset = 0x40

// ascii -> ebcdic lookup table
var ebcdicTable = []byte{
	0040, 0240, 0241, 0242, 0243, 0244, 0245, 0246,
	0247, 0250, 0325, 0056, 0074, 0050, 0053, 0174,
	0046, 0251, 0252, 0253, 0254, 0255, 0256, 0257,
	0260, 0261, 0041, 0044, 0052, 0051, 0073, 0176,
	0055, 0057, 0262, 0263, 0264, 0265, 0266, 0267,
	0270, 0271, 0313, 0054, 0045, 0137, 0076, 0077,
	0272, 0273, 0274, 0275, 0276, 0277, 0300, 0301,
	0302, 0140, 0072, 0043, 0100, 0047, 0075, 0042,
	0303, 0141, 0142, 0143, 0144, 0145, 0146, 0147,
	0150, 0151, 0304, 0305, 0306, 0307, 0310, 0311,
	0312, 0152, 0153, 0154, 0155, 0156, 0157, 0160,
	0161, 0162, 0136, 0314, 0315, 0316, 0317, 0320,
	0321, 0345, 0163, 0164, 0165, 0166, 0167, 0170,
	0171, 0172, 0322, 0323, 0324, 0133, 0326, 0327,
	0330, 0331, 0332, 0333, 0334, 0335, 0336, 0337,
	0340, 0341, 0342, 0343, 0344, 0135, 0346, 0347,
	0173, 0101, 0102, 0103, 0104, 0105, 0106, 0107,
	0110, 0111, 0350, 0351, 0352, 0353, 0354, 0355,
	0175, 0112, 0113, 0114, 0115, 0116, 0117, 0120,
	0121, 0122, 0356, 0357, 0360, 0361, 0362, 0363,
	0134, 0237, 0123, 0124, 0125, 0126, 0127, 0130,
	0131, 0132, 0364, 0365, 0366, 0367, 0370, 0371,
	0060, 0061, 0062, 0063, 0064, 0065, 0066, 0067,
	0070, 0071, 0372, 0373, 0374, 0375, 0376, 0377,
}

// variables used in xxd*()
var (
	dumpType int

	space        = []byte(" ")
	doubleSpace  = []byte("  ")
	dot          = []byte(".")
	newLine      = []byte("\n")
	zeroHeader   = []byte("0000000: ")
	unsignedChar = []byte("unsigned char ")
	unsignedInt  = []byte("};\nunsigned int ")
	lenEquals    = []byte("_len = ")
	brackets     = []byte("[] = {")
	asterisk     = []byte("*")
	commaSpace   = []byte(", ")
	comma        = []byte(",")
	semiColonNl  = []byte(";\n")
	bar          = []byte("|")
)

type XxdConfig struct {
	DumpType   int
	AutoSkip   bool
	Bars       bool
	Binary     bool
	Columns    int
	Ebcdic     bool
	Group      int
	Cfmt       bool
	Length     int
	Postscript bool
	Reverse    bool
	Seek       string
	Upper      bool
	Version    bool

	//	var (
	//	//autoskip   = flag.BoolP("autoskip", "a", false, "toggle autoskip (* replaces nul lines")
	//	//bars       = flag.BoolP("bars", "B", false, "print |ascii| instead of ascii")
	//	//binary     = flag.BoolP("binary", "b", false, "binary dump, incompatible with -ps, -i, -r")
	//	//columns    = flag.IntP("cols", "c", -1, "format <cols> octets per line")
	//	//ebcdic     = flag.BoolP("ebcdic", "E", false, "use EBCDIC instead of ASCII")
	//	//group      = flag.IntP("group", "g", -1, "num of octets per group")
	//	//cfmt       = flag.BoolP("include", "i", false, "output in C include format")
	//	//length     = flag.Int64P("len", "l", -1, "stop after len octets")
	//	//postscript = flag.BoolP("ps", "p", false, "output in postscript plain hd style")
	//	//reverse    = flag.BoolP("reverse", "r", false, "convert hex to binary")
	//	//seek       = flag.StringP("seek", "s", "", "start at seek bytes abs")
	//	upper      = flag.BoolP("uppercase", "u", false, "use uppercase hex letters")
	//	version    = flag.BoolP("version", "v", false, "print version")
	//
	// )
}

type XxdOption func(cfg *XxdConfig)

func WithEbcdic(cfg *XxdConfig) {
	cfg.Ebcdic = true
}

//TODO:: Support other options
