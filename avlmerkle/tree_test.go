package avlmerkle

import (
	"testing"
	"xfsgo/common"
	"xfsgo/test"
)

func Test_Iterator(t *testing.T) {
	db := test.NewMemStorage()
	tree, err := NewEpochTrie(common.Hash{}, db)
	if err != nil {
		t.Error(err)
		return
	}
}
