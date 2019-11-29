package nodes
import (
	"math/rand"
)

func RandomRange(k, n int) []int{
	var indexes []int
	for i := 0; i < n; i++{
		indexes = append(indexes, i)
	}
	var randomIndexes []int
	for i := 0; i < k; i++{
		randomIndexes = append(randomIndexes, indexes[rand.Intn(len(indexes))])
	}
	return randomIndexes
}		