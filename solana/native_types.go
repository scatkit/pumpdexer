package solana
import(
  "encoding/json"
  "fmt"
  "github.com/mr-tron/base58"
  "encoding/base64"
)

// Solana's data field 
type Data struct{
  Content  []byte
  Encoding EncodingType
}

func (t Data) MarshalJSON() ([]byte, error){
  return json.Marshal([]interface{}{
    t.String(),
    t.Encoding,
  })
}

func (t Data) String() string{
  switch EncodingType(t.Encoding){
  case EncodingBase58:
    return base58.Encode(t.Content)
  case EncodingBase64:
    return base64.StdEncoding.EncodeToString(t.Content)
  default:
    // TODO
    return ""
  }
}

//   UnmarshalJSON <-- method to define on custom types to customize how Json is unmarshalled into that type.
func (t *Data) UnmarshalJSON(data []byte) (err error){
  var input []string
  if err := json.Unmarshal(data, &input); err != nil{
    return err
  }
  if len(input) != 2{
    return fmt.Errorf("invalid length for Solana data, exptected 2, got %d", len(input))
  }
  contentString := input[0]
  encodingString := input[1]
  t.Encoding = EncodingType(encodingString)
  
  if contentString == ""{
    t.Content = []byte{}
    return nil
  } 
  
  switch t.Encoding{
  case EncodingBase58:
    var err error
    t.Content, err = base58.Decode(contentString)
    if err != nil{
      return err
    }
  case EncodingBase64:
    var err error
    t.Content, err = base64.StdEncoding.DecodeString(contentString)
    if err != nil{
      return err
    }
  default:
    return fmt.Errorf("unsupported encoding %s", encodingString)
  }
  return 
}

type Signature [64]byte


type EncodingType string

const (
  EncodingBase58 EncodingType = "base58"
  EncodingBase64 EncodingType = "base64"
  EncodingBase64Zstd EncodingType = "base64+Zstd"
  EncodingJsonParsed EncodingType = "jsonParsed"
  EncodingJSON EncodingType = "json"
)
