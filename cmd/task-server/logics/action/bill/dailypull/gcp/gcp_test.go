package gcp

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestXxx(t *testing.T) {
	var d decimal.Decimal
	d = d.Add(decimal.NewFromFloat(-4))
	t.Errorf("%s", d.String())
}
