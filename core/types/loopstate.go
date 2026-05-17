package types

import "errors"

// LoopState holds a Loop state used in for/next loops.
//
// Code is a pointer so LoopState doesn't transitively embed Algorithm's
// sync.Mutex — that previously made LoopState unsafe to copy and led to
// silent lock-by-value bugs whenever LoopState was stored in a map or
// slice and a method on the copied Algorithm tried to take a lock that
// served no synchronization purpose.
type LoopState struct {
	Step    float64
	Start   float64
	Finish  float64
	VarName string
	Code    *Algorithm
	Entry   CodeRef
}

// MarshalBinary packs the structure into []uint64
func (ls *LoopState) MarshalBinary() ([]uint64, error) {
	data := PackName(ls.VarName, 16)

	data = append(data, Float2uint(float32(ls.Step)))
	data = append(data, Float2uint(float32(ls.Start)))
	data = append(data, Float2uint(float32(ls.Finish)))
	data = append(data, uint64(ls.Entry.Line))
	data = append(data, uint64(ls.Entry.Statement))

	return data, nil
}

// UnmarshalBinary loads the structure for []uint64
func (ls *LoopState) UnmarshalBinary(data []uint64) error {

	if len(data) < 9 {
		return errors.New("Not enough data to unpack loop state")
	}

	// ok got enough data to decode
	ls.VarName = UnpackName(data[0:4])
	ls.Step = float64(Uint2Float(data[4]))
	ls.Start = float64(Uint2Float(data[5]))
	ls.Finish = float64(Uint2Float(data[6]))
	ls.Entry.Line = int(data[7])
	ls.Entry.Statement = int(data[8])

	return nil
}
