package qingcloud

import (
	"fmt"
)

func validateInstanceType() func(v interface{}, k string) (ws []string, errors []error) {
	var validate = map[string][]string{
		"pek1": []string{"c1m1", "c1m2", "c1m4", "c2m2", "c2m4", "c2m8", "c4m4", "c4m8", "c4m16"},
		"else": []string{"small_b", "small_c", "medium_a", "medium_b", "medium_c", "large_a", "large_b", "large_c"},
	}
	var limitsMap = make(map[string]bool)
	for _, v := range limits {
		limitsMap[v] = true
	}

	return func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)
		if limitsMap[value] {
			return
		}
		errors = append(errors, fmt.Errorf("%q (%q) doesn't match  %q", k, value))
		return
	}
}
