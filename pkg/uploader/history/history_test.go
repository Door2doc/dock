package history

import (
	"reflect"
	"testing"
)

func TestHistory_Events(t *testing.T) {
	t.Run("two events", func(t *testing.T) {
		h := New()
		e1 := h.NewEvent()
		e2 := h.NewEvent()

		if e1 == e2 {
			t.Fatal("did not create new event")
		}
		want := []*Event{e2, e1}
		got := h.Events()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Events() == %v, got %v", want, got)
		}
	})

	t.Run("eleven events", func(t *testing.T) {
		h := New()
		var want []*Event
		for i := 0; i <= MaxHistory; i++ {
			e := h.NewEvent()
			if i == 0 {
				// first event should be filtered out
				continue
			}
			want = append([]*Event{e}, want...)
		}

		got := h.Events()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Events() == \n\t%v, got \n\t%v", want, got)
		}
	})

	t.Run("lots and lots of events", func(t *testing.T) {
		h := New()
		for i := 0; i < MaxHistory*1000; i++ {
			h.NewEvent()
		}
		got := h.Events()

		if len(got) != MaxHistory {
			t.Errorf("len() == %d, got %d", MaxHistory, len(got))
		}
		if cap(got) != MaxHistory {
			t.Errorf("cap() == %d, got %d", MaxHistory, cap(got))
		}

		// testing the internals
		if len(h.events) != MaxHistory {
			t.Errorf("len() == %d, got %d", MaxHistory, len(h.events))
		}
		if cap(h.events) > MaxHistory*2 {
			t.Errorf("cap() <= %d, got %d", MaxHistory*2, cap(h.events))
		}
	})
}
