//go:build !solution

package testequal

import "fmt"

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are equal.
func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	if !deepEqual(expected, actual) {
		t.Errorf(formatMsg(msgAndArgs...))
		return false
	}
	return true
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	if deepEqual(expected, actual) {
		t.Errorf(formatMsg(msgAndArgs...))
		return false
	}
	return true
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertNotEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

func deepEqual(expected, actual interface{}) bool {
	switch exp := expected.(type) {
	case int:
		act, ok := actual.(int)
		return ok && exp == act
	case int16:
		act, ok := actual.(int16)
		return ok && exp == act
	case int32:
		act, ok := actual.(int32)
		return ok && exp == act
	case int64:
		act, ok := actual.(int64)
		return ok && exp == act
	case int8:
		act, ok := actual.(int8)
		return ok && exp == act
	case uint:
		act, ok := actual.(uint)
		return ok && exp == act
	case uint16:
		act, ok := actual.(uint16)
		return ok && exp == act
	case uint32:
		act, ok := actual.(uint32)
		return ok && exp == act
	case uint64:
		act, ok := actual.(uint64)
		return ok && exp == act
	case uint8:
		act, ok := actual.(uint8)
		return ok && exp == act

	case string:
		act, ok := actual.(string)
		return ok && exp == act

	case map[string]string:
		act, ok := actual.(map[string]string)
		if !ok || len(exp) != len(act) {
			return false
		}
		if exp == nil || act == nil {
			return exp == nil && act == nil
		}
		for expK, expV := range exp {
			if actV, ok := act[expK]; !ok || expV != actV {
				return false
			}
		}
		return true

	case []int:
		act, ok := actual.([]int)
		if !ok || len(exp) != len(act) {
			return false
		}
		if exp == nil || act == nil {
			return exp == nil && act == nil
		}
		for i := range exp {
			if exp[i] != act[i] {
				return false
			}
		}
		return true

	case []byte:
		act, ok := actual.([]byte)
		if !ok || len(exp) != len(act) {
			return false
		}
		if exp == nil || act == nil {
			return exp == nil && act == nil
		}
		for i := range exp {
			if exp[i] != act[i] {
				return false
			}
		}
		return true

	default:
		return false
	}
}

func formatMsg(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	if format, ok := msgAndArgs[0].(string); ok && len(msgAndArgs) > 1 {
		return fmt.Sprintf(format, msgAndArgs[1:]...)
	}
	return fmt.Sprint(msgAndArgs[0])
}
