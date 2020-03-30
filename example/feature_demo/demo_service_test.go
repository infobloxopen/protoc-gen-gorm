package example

import (
	"reflect"
	"testing"
)

func TestMultipleCreate(t *testing.T) {
	t.Run("... multi ", func(t *testing.T) {
		service := NewMultipleCreatesClient(nil)
		ref := reflect.TypeOf(service)
		method1 := "CreateA"
		method2 := "CreateB"
		method3 := "CreateC"
		_, ok := ref.MethodByName(method1)
		if !ok {
			t.Errorf("Method %s doesn't exist", method1)
		}
		_, ok = ref.MethodByName(method2)
		if !ok {
			t.Errorf("Method %s doesn't exist", method2)
		}
		_, ok = ref.MethodByName(method3)
		if !ok {
			t.Errorf("Method %s doesn't exist", method3)
		}
	})
}
