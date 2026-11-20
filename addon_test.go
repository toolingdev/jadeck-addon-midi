package main

import (
	"testing"

	"github.com/toolingdev/ptz-controller/sdk/addon/addontest"
)

func TestOverThePipe(t *testing.T) {
	addontest.Start(t, spec()) // manifest valid and it answers initialize
}
