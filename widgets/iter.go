package widgets

import (
	"iter"
	"reflect"
)

// seqRanger adapts iter.Seq[T] to Jet's Ranger interface for use in templates.
// Jet requires Ranger with Range() (reflect.Value, reflect.Value, bool) signature.
type seqRanger[T any] struct {
	next func() (T, bool)
	stop func()
	idx  int
}

func newSeqRanger[T any](seq iter.Seq[T]) *seqRanger[T] {
	next, stop := iter.Pull(seq)

	return &seqRanger[T]{
		next: next,
		stop: stop,
		idx:  -1,
	}
}

func (r *seqRanger[T]) Range() (index, value reflect.Value, end bool) {
	val, ok := r.next()
	if !ok {
		r.stop()

		return reflect.Value{}, reflect.Value{}, true
	}

	r.idx++

	return reflect.ValueOf(r.idx), reflect.ValueOf(val), false
}

func (r *seqRanger[T]) ProvidesIndex() bool {
	return true
}
