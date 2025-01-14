package utils

import "github.com/saiset-co/sai-eth-interaction/models"

func RemoveContract(slice []models.Contract, s int) []models.Contract {
	return append(slice[:s], slice[s+1:]...)
}

func GetMaxKey(m map[uint64]bool) uint64 {
	var maxKey uint64
	for k := range m {
		if k > maxKey {
			maxKey = k
		}
	}
	return maxKey
}
