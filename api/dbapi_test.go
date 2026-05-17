package s8webclient

import "testing"

// TestInsert is a pre-existing stub for a future dbapi insert test. It was
// previously `t.Fail()` without a body, meaning the package never passed.
// Skipping until the test is fleshed out — the body needs a fake DuckTape
// server to exercise the request/response cycle.
func TestInsert(t *testing.T) {
	t.Skip("dbapi insert test not yet implemented; requires a fake DuckTape server")
	//Do("insert into frogs (age,count) values (5,5)")
}
