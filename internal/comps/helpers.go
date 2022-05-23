package comps

import (
	"crypto/rand"
	log "github.com/sirupsen/logrus"
	"math/big"
)

func genRandomBoundary(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letter))))
		if err != nil {
			log.Fatalf("can't generate boundary - %v", err)
		}
		b[i] = letter[n.Uint64()]
	}
	return string(b)
}
