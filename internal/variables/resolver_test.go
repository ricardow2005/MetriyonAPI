package variables

import "testing"

func TestMergePriorityAndResolve(t *testing.T) {
	values := Merge(map[string]string{"host": "global", "onlyGlobal": "yes"}, map[string]string{"host": "environment"}, map[string]string{"host": "collection"}, map[string]string{"host": "request"})
	got, missing := Resolve("https://{{host}}/{{onlyGlobal}}/{{missing}}", values)
	if got != "https://request/yes/{{missing}}" {
		t.Fatalf("unexpected result: %s", got)
	}
	if len(missing) != 1 || missing[0] != "missing" {
		t.Fatalf("unexpected missing variables: %#v", missing)
	}
}
