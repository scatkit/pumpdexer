package system
import(
  "bytes"
  "fmt"
  
  bin "github.com/gagliardetto/binary"
)

func encodeT(data interface{}, buf *bytes.Buffer) error {
	if err := bin.NewBinEncoder(buf).Encode(data); err != nil {
		return fmt.Errorf("unable to encode instruction: %w", err)
	}
	return nil
}

func decodeT(dst interface{}, data []byte) error {
	return bin.NewBinDecoder(data).Decode(dst)
}
