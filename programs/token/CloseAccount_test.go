package token
import(
  "testing"
  "strconv"
  "bytes"
  
  "github.com/gagliardetto/gofuzz"
  "github.com/stretchr/testify/require"
  //"github.com/davecgh/go-spew/spew"
)

func TestEncodeDecode_CloseAccount(t *testing.T){
  fz := fuzz.New().NilChance(0)
  for i:=0; i<1; i++{
    t.Run("CloseAccount"+strconv.Itoa(i), func(t *testing.T){
      {
        params := new(CloseAccount)
        fz.Fuzz(params)
        params.Accounts = nil
        params.Signers = nil
        buf := new(bytes.Buffer)
        err := encodeT(*params, buf)
        require.NoError(t, err)
        // 
        got := new(CloseAccount)
        err = decodeT(got, buf.Bytes())
        got.Accounts = nil
        got.Signers = nil
        require.NoError(t, err)
        require.Equal(t, params, got) 
      }
    })
  }
}

