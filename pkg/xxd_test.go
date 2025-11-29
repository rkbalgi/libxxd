package xxd

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

func TestXXD(t *testing.T) {

	xxdCfg := &XxdConfig{}
	xxdCfg.DumpType = DumpPostscript

	inFile, err := os.Open("./testdata/hello.txt")
	if err != nil {
		t.Error(err)
	}

	buf := &bytes.Buffer{}

	if err := xxd(inFile, buf, "-", xxdCfg); err != nil {
		t.Error(err)
	}
	t.Log(hex.EncodeToString(buf.Bytes()))
}
