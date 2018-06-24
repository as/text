package kbd

import (
	"testing"

	"github.com/as/text"
	"golang.org/x/mobile/event/key"
)

func TestSendClientContW(t *testing.T) {
	ed, err := text.Open(text.NewBuffer())
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	have := []byte("The quick brown fox")
	want := []byte("The quick ")
	ed.Insert(have, 0)
	ed.Select(ed.Len(), ed.Len())
	SendClient(ed, key.Event{Rune: rune(17), Direction: 1})
	SendClient(ed, key.Event{Rune: rune(17), Direction: 1})
	SendClient(ed, key.Event{Rune: rune(17), Direction: 1})
	have = ed.Bytes()
	if string(have) != string(want) {
		t.Logf("bad index: have=%q want=%q\n", have, want)
	}
}
