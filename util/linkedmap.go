package util

type LinkedMap[K comparable, V any] struct {
	m    map[K]V
	keys []K
}

func NewLinkedMap[K comparable, V any]() *LinkedMap[K, V] {
	return &LinkedMap[K, V]{
		m:    make(map[K]V),
		keys: make([]K, 0),
	}
}

func (lm *LinkedMap[K, V]) Set(key K, value V) {
	if _, ok := lm.m[key]; ok {
		lm.m[key] = value
		return
	}
	lm.m[key] = value
	lm.keys = append(lm.keys, key)
}

func (lm *LinkedMap[K, V]) Get(key K) (V, bool) {
	v, ok := lm.m[key]
	return v, ok
}

func (lm *LinkedMap[K, V]) GetIndex(i int) (V, bool) {
	if i < 0 || i >= lm.Len() {
		var v V
		return v, false
	}

	return lm.m[lm.keys[i]], true
}

func (lm *LinkedMap[K, V]) Len() int {
	return len(lm.keys)
}

func (lm LinkedMap[K, V]) Iterate() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		for _, key := range lm.keys {
			if !yield(key, lm.m[key]) {
				return
			}
		}
	}
}
