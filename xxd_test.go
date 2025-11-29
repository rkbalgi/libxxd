package xxd_test

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

	fmt.Println(inFile)

	data, _ := ioutil.ReadAll(inFile)
	t.Log(hex.EncodeToString(data))

	b := &bytes.Buffer{}
	buf := bufio.NewWriter(b)
	//println(buf.Size())
	xxdCfg := &xxd.XxdConfig{DumpType: 0, AutoSkip: false, Bars: true, Binary: false, Columns: -1, Ebcdic: false, Group: -1, Cfmt: false, Length: -1, Postscript: false, Reverse: false, Seek: "", Upper: false, Version: false}
	//xxdCfg.DumpType = xxd.DumpPostscript

	if err := xxd.Xxd(inFile, buf, fileName, xxdCfg); err != nil {
		t.Error(err)
	}

	println(buf.Size())
	t.Log(b)
}
