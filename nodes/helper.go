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
		ridx :=rand.Intn(len(indexes))
		randomIndexes = append(randomIndexes, indexes[ridx])
		indexes = append(indexes[:ridx], indexes[ridx + 1:]...)

	}
	return randomIndexes
}		 


func Contains(slice []string, e string) bool {
	for _, s := range slice {
		if s == e {
			return true
		}
	}
	return false
} 