package token

import (
	"bytes"
	 "github.com/gagliardetto/gofuzz"
   bin "github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestEncodeDecode_InitializeAccount(t *testing.T) {
	fu := fuzz.New().NilChance(0)
	for i := 0; i < 1; i++ {
		t.Run("InitializeAccount"+strconv.Itoa(i), func(t *testing.T) {
			{
				params := new(InitializeAccount)
				fu.Fuzz(params)
				params.AccountMetaSlice = nil
				buf := new(bytes.Buffer)
				err := encodeT(*params, buf)
				bin.NoError(t, err)
				//
				got := new(InitializeAccount)
				err = decodeT(got, buf.Bytes())
				got.AccountMetaSlice = nil
				bin.NoError(t, err)
				bin.Equal(t, params, got)
			}
		})
	}
}
