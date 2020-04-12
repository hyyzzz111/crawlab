// Code generated by "enumer -type EnvMode -linecomment"; DO NOT EDIT.

package runtime

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the enumer command to generate them again.
	var x [1]struct{}
	_ = x[Development-0]
	_ = x[Production-1]
	_ = x[Test-2]
}
func (i EnvMode) Values() []string {
	return []string{
		"Development",
		"Production",
		"Test",
	}
}

var _EnvMode_kv = map[string]int64{
	"Development": 0,
	"Production":  1,
	"Test":        2,
}

func (i EnvMode) KV() map[string]int64 {
	return _EnvMode_kv
}

var _EnvMode_vk = map[int64]string{
	0: "Development",
	1: "Production",
	2: "Test",
}

func (i EnvMode) VK() map[int64]string {
	return _EnvMode_vk
}

const _EnvMode_name = "DevelopmentProductionTest"

var _EnvMode_index = [...]uint8{0, 11, 21, 25}

func (i EnvMode) String() string {
	if i < 0 || i >= EnvMode(len(_EnvMode_index)-1) {
		return "EnvMode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EnvMode_name[_EnvMode_index[i]:_EnvMode_index[i+1]]
}
