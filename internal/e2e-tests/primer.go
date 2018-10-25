package e2etests

type Primer struct {
	val uint
}

func NewPrimer() *Primer {
	return &Primer{val: 2}
}

func (c *Primer) Next() uint {
	for isPrime(c.val) {
		c.val++
	}

	next := c.val
	c.val++
	return next
}

func isPrime(val uint) bool {
	for i := uint(2); i < val; i++ {
		if val%i == 0 {
			return false
		}
	}

	return true
}
