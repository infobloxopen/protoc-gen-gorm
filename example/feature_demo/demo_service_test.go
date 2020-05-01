package example

import (
	fmt "fmt"
	"reflect"
	"testing"
)

func TestMultipleCrud(t *testing.T) {
	t.Run("Multi crud methods", func(t *testing.T) {
		service := NewMultipleMethodsAutoGenClient(nil)
		ref := reflect.TypeOf(service)

		methods := []string{"Create", "Read", "Update", "List", "Delete", "DeleteSet"}
		repeated := []string{"A", "B"}

		for idx := 0; idx < len(methods); idx++ {
			for i := 0; i < len(repeated); i++ {
				methodName := fmt.Sprintf("%s%s", methods[idx], repeated[i])
				_, ok := ref.MethodByName(methodName)
				if !ok {
					t.Errorf("Method %s doesn't exist", methodName)
				}
			}
		}
	})
}
