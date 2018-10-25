package main

import (
	"github.com/aphistic/sweet"
	_ "github.com/efritz/go-mockgen/matchers"
	. "github.com/onsi/gomega"
)

type E2ESuite struct{}

func (s *E2ESuite) TestCalls(t sweet.T) {
	// TODO
	Expect(true).To(BeTrue())
}
