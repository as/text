package text

import "testing"

func TestMultiDel(t *testing.T) {
	tbl := [][3][2]int64{
		{{2, 9}, {1, 5}, {1, 1}},
		{{2, 9}, {3, 8}, {1, 1}},
		{{2, 9}, {7, 14}, {2, 6}},
	}
	for _, v := range tbl {
		q0, q1 := Coherence(-1, v[0][0], v[0][1], v[1][0], v[1][1])
		if q0 != v[2][0] || q1 != v[2][1] {
			t.Fail()
		}
	}
}
