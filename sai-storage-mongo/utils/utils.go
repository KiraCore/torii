package utils

func MapsDelete[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	for k, _ := range src {
		delete(dst, k)
	}
}
