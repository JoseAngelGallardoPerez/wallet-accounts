package event

import "github.com/olebedev/emitter"

var evEmitter = emitter.New(5)

func Emitter() *emitter.Emitter {
	return evEmitter
}
