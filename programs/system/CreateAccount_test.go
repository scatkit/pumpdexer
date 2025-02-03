package system
import(
  "testing"
  "strconv"
  "bytes"
  
  "github.com/gagliardetto/gofuzz"
  "github.com/stretchr/testify/require"
)

func TestEncodeDecode_CreateAccount(t *testing.T){
  fz := fuzz.New().NilChance(0)
  for i:=0; i<1; i++{
    t.Run("CreateAccount"+strconv.Itoa(i), func(t *testing.T){
      params := new(CreateAccount)
      fz.Fuzz(params)
      params.AccountMetaSlice = nil
      buf := new(bytes.Buffer)
      err := encodeT(*params, buf)
      require.NoError(t, err)
      got := new(CreateAccount)
      err = decodeT(got, buf.Bytes())
      got.AccountMetaSlice = nil
      require.NoError(t, err)
      require.Equal(t, params, got) 
    })
  }
}
