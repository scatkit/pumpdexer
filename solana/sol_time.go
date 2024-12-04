package solana
import "time"

type UnixTimeSeconds int64
 
func (ut UnixTimeSeconds) Time() time.Time{
  return time.Unix(int64(ut),0) // expects to have arguments: int64, 0 -> nanoseconds
}
