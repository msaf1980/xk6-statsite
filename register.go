package statsite

import (
	"go.k6.io/k6/output"
)

func init() {
	output.RegisterExtension("statsite", New)
}
