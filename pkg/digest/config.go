package digest

import "io"

// Config represents configurations that can be passed
// to create a Digest.
//
// Key: The primary key positions
// Value: The Value positions that needs to be compared for diff
// Include: Include these positions in output. It is Value positions by default.
type Config struct {
	Key                Positions
	Value              Positions
	Include            Positions
	Reader             io.Reader
	Separator          rune
	LazyQuotes         bool
	IgnoreColumnsCheck bool // 追加
	RawSplit           bool // 追加
}

// NewConfig creates an instance of Config struct.
func NewConfig(
	r io.Reader,
	primaryKey Positions,
	valueColumns Positions,
	includeColumns Positions,
	separator rune,
	lazyQuotes bool,
	ignoreColumnsCheck bool, // 追加
	rawSplit bool, // 追加
) *Config {
	if len(includeColumns) == 0 {
		includeColumns = valueColumns
	}

	return &Config{
		Reader:             r,
		Key:                primaryKey,
		Value:              valueColumns,
		Include:            includeColumns,
		Separator:          separator,
		LazyQuotes:         lazyQuotes,
		IgnoreColumnsCheck: ignoreColumnsCheck, // 追加
		RawSplit:           rawSplit,           // 追加
	}
}
