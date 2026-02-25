package digest

import (
	"fmt"
	"runtime"
)

type messageType int

const (
	addition     messageType = iota
	modification messageType = iota
	deletion     messageType = iota
	unchanged    messageType = iota // 追加
)

// Differences represents the differences
// between 2 csv content
type Differences struct {
	Additions     []Addition
	Modifications []Modification
	Deletions     []Deletion
	Unchanged     []Unchanged // 追加
}

// Addition is a row appearing in delta but missing in base
type Addition []string

// Deletion is a row appearing in base but missing in delta
type Deletion []string

// Modification is a row present in both delta and base
// with the values column changed in delta
type Modification struct {
	Original []string
	Current  []string
}

// Unchanged is a row present in both base and delta without changes
type Unchanged []string // 追加

type message struct {
	original []string
	current  []string
	_type    messageType
}

// Diff finds the Differences between baseConfig and deltaConfig
func Diff(baseConfig, deltaConfig Config) (Differences, error) {
	baseEngine := NewEngine(baseConfig)
	baseDigestChannel, baseErrorChannel := baseEngine.StreamDigests()

	baseFileDigest := NewFileDigest()
	for digests := range baseDigestChannel {
		for _, d := range digests {
			baseFileDigest.Append(d)
		}
	}

	if err := <-baseErrorChannel; err != nil {
		return Differences{}, fmt.Errorf("error processing base file: %v", err)
	}

	deltaEngine := NewEngine(deltaConfig)
	deltaDigestChannel, deltaErrorChannel := deltaEngine.StreamDigests()

	additions := make([]Addition, 0)
	modifications := make([]Modification, 0)
	deletions := make([]Deletion, 0)
	unchangedRows := make([]Unchanged, 0) // 追加

	msgChannel := streamDifferences(baseFileDigest, deltaDigestChannel)
	for msg := range msgChannel {
		switch msg._type {
		case addition:
			additions = append(additions, msg.current)
		case modification:
			modifications = append(modifications, Modification{Original: msg.original, Current: msg.current})
		case deletion:
			deletions = append(deletions, msg.current)
		case unchanged: // 追加
			unchangedRows = append(unchangedRows, msg.current)
		default:
			continue
		}
	}

	if err := <-deltaErrorChannel; err != nil {
		return Differences{}, fmt.Errorf("error processing delta file: %v", err)
	}

	return Differences{Additions: additions, Modifications: modifications, Deletions: deletions, Unchanged: unchangedRows}, nil
}

func streamDifferences(baseFileDigest *FileDigest, digestChannel chan []Digest) chan message {
	maxProcs := runtime.NumCPU()
	msgChannel := make(chan message, maxProcs*bufferSize)

	go func(base *FileDigest, digestChannel chan []Digest, msgChannel chan message) {
		defer close(msgChannel)

		for digests := range digestChannel {
			for _, d := range digests {
				if baseValue, present := base.Digests[d.Key]; present {
					if baseValue != d.Value {
						// Modification
						msgChannel <- message{_type: modification, current: d.Source, original: base.SourceMap[d.Key]}
					} else {
						// Unchanged (追加)
						msgChannel <- message{_type: unchanged, current: d.Source}
					}
					// delete from sourceMap so that at the end only deletions are left in base
					delete(base.SourceMap, d.Key)
				} else {
					// Addition
					msgChannel <- message{_type: addition, current: d.Source}
				}
			}
		}

		for _, value := range base.SourceMap {
			msgChannel <- message{_type: deletion, current: value}
		}

	}(baseFileDigest, digestChannel, msgChannel)

	return msgChannel
}
