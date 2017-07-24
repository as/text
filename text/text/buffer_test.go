package text

import "testing"

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
