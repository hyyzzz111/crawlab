// Code generated by "stringer -type BusEvent -linecomment"; DO NOT EDIT.

package runtime

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ModeChange-0]
}

const _BusEvent_name = "ModeChange"

var _BusEvent_index = [...]uint8{0, 10}

func (i BusEvent) String() string {
	if i < 0 || i >= BusEvent(len(_BusEvent_index)-1) {
		return "BusEvent(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _BusEvent_name[_BusEvent_index[i]:_BusEvent_index[i+1]]
}
