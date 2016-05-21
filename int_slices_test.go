package parameters

import (
	"reflect"
	"testing"
)

func TestUniqueUint64(t *testing.T) {
	one := []uint64{3, 2, 1}
	if !reflect.DeepEqual(UniqueUint64(one), one) {
		t.Errorf("slice with no dupes is different")
	}
    
    two := []uint64{3, 2, 1, 3, 3, 3, 3}
    
    if !reflect.DeepEqual(UniqueUint64(two), one) {
        t.Errorf("slice with dupes is not what we expected")
    }
}
