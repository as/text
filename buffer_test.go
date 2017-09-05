package text

import "testing"

func sid(buf Buffer, q0, q1 int64) string {
	return string(buf.Bytes())
}
func i(buf Buffer, in string, q0 int64) (n int) {
	return buf.Insert([]byte(in), q0)
}
func fail(t *testing.T, name, have, want string) {
	t.Logf("%s: have: %q\n\t\twant: %q\n", name, have, want)
	t.Fail()
}
func want(t *testing.T, b Buffer, want string) {
	if h := sid(b, 0, 99); h != want {
		fail(t, "TestInsert", h, want)
	}
}

func TestInsert(t *testing.T) {
	b := NewBuffer()
	i(b, "test", 0)
	if h := sid(b, 0, 99); h != "test" {
		fail(t, "TestInsert", h, "test")
	}
}

func TestInsert3X(t *testing.T) {
	b := NewBuffer()
	i(b, "test", 0)
	i(b, "a ", 0)
	i(b, " for", 9999)
	want(t, b, "a test for")

}

func BenchmarkInsert1K(b *testing.B) {
	buf := NewBuffer()
	input := []byte("0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf.Insert(input, 0)
	}

	b.StopTimer()
}

func BenchmarkDelete1K(b *testing.B) {
	buf := NewBuffer()
	input := []byte("0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")
	for i := 0; i < b.N; i++ {
		buf.Insert(input, 0)
	}
	l := int64(len(input))
	b.StartTimer()
	for buf.Len() != 0 {
		buf.Delete(0, l)
	}
	b.StopTimer()
}
