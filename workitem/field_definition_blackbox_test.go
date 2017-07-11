package workitem_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/fabric8-services/fabric8-auth/resource"
	. "github.com/fabric8-services/fabric8-auth/workitem"
)

func TestListFieldDefMarshalling(t *testing.T) {
	t.Parallel()
	resource.Require(t, resource.UnitTest)
	def := FieldDefinition{
		Required:    true,
		Label:       "Salt",
		Description: "Put it in your soup",
		Type: ListType{
			SimpleType:    SimpleType{Kind: KindList},
			ComponentType: SimpleType{Kind: KindString},
		},
	}
	bytes, err := json.Marshal(def)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	fmt.Printf("bytes are " + string(bytes))
	unmarshalled := FieldDefinition{}
	json.Unmarshal(bytes, &unmarshalled)

	if !reflect.DeepEqual(def, unmarshalled) {
		t.Errorf("field should be %v, but is %v", def, unmarshalled)
	}
}
