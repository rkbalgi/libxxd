package xxd_test

import (
	"bufio"
	"bytes"
	"fmt"

	"os"
	"testing"

	xxd "github.com/rkbalgi/libxxd"
)

func TestXXD(t *testing.T) {

	fileName := "./testdata/hello.txt"

	inFile, err := os.Open(fileName)
	if err != nil {
		t.Error(err)
	}

	buf := &bytes.Buffer{}
	writer := bufio.NewWriter(buf)

	xxdCfg := &xxd.XxdConfig{DumpType: 0, AutoSkip: false, Bars: true, Binary: false, Columns: -1, Ebcdic: false, Group: 16, Cfmt: false, Length: -1, Postscript: false, Reverse: false, Seek: "", Upper: false, Version: false}

	if err := xxd.XxdBasic(inFile, writer, xxdCfg); err != nil {
		t.Error(err)
	}
	writer.Flush()
	expectedLen := 495
	if len(buf.Bytes()) != expectedLen {
		t.Fatal(fmt.Sprintf("Expected: <%d>, Got: <%d>", expectedLen, len(buf.Bytes())))

	}

}
