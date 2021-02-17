package globals

// Learnings from https://github.com/9xb/awsspec
import (
	//"github.com/aws/aws-sdk-go/aws"
	"encoding/json"
)

type (
	ConditionOperator interface {
		GetOperator() string
		GetVariable() string
		GetValue() interface{}
	}
	// PolicyDocument represents an IAM policy document
	PolicyDocument struct {
		Version   string
		ID        string
		Statement []Statement
	}

	// Statement represents an IAM statement
	// UnmarshalJSON gets called for OptSlice only, to make everything a []string
	Statement struct {
		// TODO:
		// - Handle Principal, NotPrincipal, and Condition
		SID          string
		Principal    interface{}
		NotPrincipal interface{}
		Effect       string
		Action       *OptSlice
		NotAction    *OptSlice
		Resource     *OptSlice
		NotResource  *OptSlice
		Condition    map[ConditionType]map[ConditionVariable]OptSlice `json:",omitempty"`
	}
	// OptSlice is an entity that could be either a JSON string or a slice
	// As per https://stackoverflow.com/a/38757780/543423
	OptSlice []string

	// ConditionType represents all the possible comparison types for the
	// Condition of a Policy Statement
	// Inspired by github.com/gwkunze/goiam/policy
	ConditionType string

	// ConditionVariable represent the available variables used in Conditions
	// Inspired by github.com/gwkunze/goiam/policy
	ConditionVariable string
)

// UnmarshalJSON sets *o to a copy of data
func (o *OptSlice) UnmarshalJSON(data []byte) error {
	// Use normal json.Unmarshal for subtypes
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var v []string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		//fmt.Println("1st: ", v)
		*o = v
		return nil
	}
	//fmt.Println("2nd: ", s)
	*o = []string{s}
	return nil
}

// find takes a slice and looks for an element in it.
func find(slice []string, val string) (res bool) {
	for _, item := range slice {
		if item == val {
			res = true
			return
		}
	}
	return
}

// Contains checks whether OptSlice contains the provided items slice
func (o OptSlice) Contains(items []string) (res bool) {
	if len(items) > len(o) {
		return false
	}

	for _, e := range items {
		if !find(o, e) {
			return false
		}
	}

	return true
}
