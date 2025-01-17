package solana

import (
	"encoding/base64"
	"fmt"

	"github.com/mr-tron/base58"
)

// Solana's data field
type Data struct {
	Content  []byte
	Encoding EncodingType
}

func (t Data) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		t.String(),
		t.Encoding,
	})
}

func (t Data) String() string {
	switch EncodingType(t.Encoding) {
	case EncodingBase58:
		return base58.Encode(t.Content)
	case EncodingBase64:
		return base64.StdEncoding.EncodeToString(t.Content)
	default:
		// TODO
		return ""
	}
}

// UnmarshalJSON <-- method to define on custom types to customize how Json is unmarshalled into that type.
func (t *Data) UnmarshalJSON(data []byte) (err error) {
	var input []string
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	if len(input) != 2 {
		return fmt.Errorf("invalid length for Solana data, exptected 2, got %d", len(input))
	}
	contentString := input[0]
	encodingString := input[1]
	t.Encoding = EncodingType(encodingString)

	if contentString == "" {
		t.Content = []byte{}
		return nil
	}

	switch t.Encoding {
	case EncodingBase58:
		var err error
		t.Content, err = base58.Decode(contentString)
		if err != nil {
			return err
		}
	case EncodingBase64:
		var err error
		t.Content, err = base64.StdEncoding.DecodeString(contentString)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported encoding %s", encodingString)
	}
	return
}

type Signature [64]byte

var zeroSignature = Signature{}

func SignatureFromBase58(b58Trans string) (out Signature, err error) {
	val, err := base58.Decode(b58Trans)
	if err != nil {
		return
	}

	if len(val) != SignatureLength {
		err = fmt.Errorf("invalid singature length, expected 64, got %d", len(val))
		return
	}

	copy(out[:], val)
	return out, err
}

func (sig Signature) IsZero() bool {
	return sig == zeroSignature
}

func (sig Signature) Equals(pb Signature) bool {
	return sig == pb
}

func (s Signature) String() string {
	return base58.Encode(s[:])
}

func (p *Signature) UnmarshalText(data []byte) (err error) {
	tmp, err := SignatureFromBase58(string(data))
	if err != nil {
		return fmt.Errorf("invalid signature %q: %w", string(data), err)
	}
	*p = tmp
	return
}

func (p Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(base58.Encode(p[:]))
}

func (p *Signature) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}

	dat, err := base58.Decode(s)
	if err != nil {
		return err
	}

	if len(dat) != SignatureLength {
		return fmt.Errorf("invalid length for Signature, expected 64, got %d", len(dat))
	}

	target := Signature{}
	copy(target[:], dat)
	*p = target
	return
}

type Base58 []byte

type Hash PublicKey

// Decodes a base-58 string into Hash
func MustHashFromBase58(in string) Hash {
	return Hash(MustPubkeyFromBase58(in))
}

func HashFromBase58(in string) (Hash, error) {
	tmp, err := PublicKeyFromBase58(in)
	if err != nil {
		return Hash{}, err
	}
	return Hash(tmp), nil
}

func HashFromBytes(in []byte) Hash {
	return Hash(PublicKeyFromBytes(in))
}

func (ha Hash) MarshalText() ([]byte, error) {
	s := base58.Encode(ha[:])
	return []byte(s), nil
}

func (ha *Hash) UnmarshalText(data []byte) (err error) {
	tmp, err := HashFromBase58(string(data))
	if err != nil {
		return fmt.Errorf("invalid hash %q: %w", string(data), err)
	}
	*ha = tmp
	return
}

func (ha Hash) String() string {
	return base58.Encode(ha[:])
}

func (ha Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(base58.Encode(ha[:]))
}

func (ha *Hash) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	tmp, err := HashFromBase58(s)
	if err != nil {
		return fmt.Errorf("invalid hash %q: %w", s, err)
	}
	*ha = tmp
	return
}

func (ha Hash) Equals(pb Hash) bool {
	return ha == pb
}

var zeroHash = Hash{}

func (ha Hash) IsZero() bool {
	return ha == zeroHash
}

type EncodingType string

const (
	EncodingBase58     EncodingType = "base58"
	EncodingBase64     EncodingType = "base64"
	EncodingBase64Zstd EncodingType = "base64+Zstd"
	EncodingJsonParsed EncodingType = "jsonParsed"
	EncodingJSON       EncodingType = "json"
)
