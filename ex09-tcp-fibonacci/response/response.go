package response

import (
	"math/big"
	"time"
)

type Response struct {
	Result *big.Int
	Spent  time.Duration
}
