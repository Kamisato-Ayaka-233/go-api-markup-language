package parser

type Handler[T any, V any] struct {
	cache V
	keys  []func(T, V) bool
	pres  []func(T, V)
	fs    []func(T, V)
}

func (hd *Handler[T, V]) Prepare(v func(T, V)) {
	hd.pres = append(hd.pres, v)
}

func (hd *Handler[T, V]) Add(k func(T, V) bool, v func(T, V)) {
	hd.keys = append(hd.keys, k)
	hd.fs = append(hd.fs, v)
}

func (hd *Handler[T, V]) Do(t T) {
	for _, pre := range hd.pres {
		pre(t, hd.cache)
	}
	for i, k := range hd.keys {
		if k(t, hd.cache) {
			hd.fs[i](t, hd.cache)
			break
		}
	}
}

func NewHandler[T any, V any](cache V) *Handler[T, V] {
	return &Handler[T, V]{
		cache,
		make([]func(T, V) bool, 0),
		make([]func(T, V), 0),
		make([]func(T, V), 0),
	}
}
