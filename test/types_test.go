package test

import (
	"fmt"
	"testing"

	types "github.com/mrkhay/gobank/type"
	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	acc, err := types.NewAccount("a", "b", "player", "")
	assert.Nil(t, err)

	fmt.Printf("%+v\n", acc)

}
