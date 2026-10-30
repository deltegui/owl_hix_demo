package entry

import (
	"testing"
	"time"

	"github.com/deltegui/owl/v3/pagination"
)

func find(t *testing.T, f entryFilter, o entryOrder, page int) ([]int, int, int64) {
	t.Helper()
	list, sum, err := newEntryService().Find(f, o, pagination.Pagination{CurrentPage: page})
	if err != nil {
		t.Fatal(err)
	}
	var ids []int
	for _, e := range list.Items {
		ids = append(ids, e.ID)
	}
	return ids, list.Pagination.TotalElements, sum
}

func TestFindPaginatesAndSums(t *testing.T) {
	ids, total, sum := find(t, entryFilter{}, entryOrder{By: "id"}, 1)
	if total != 40 || len(ids) != pageSize || ids[0] != 1 || ids[7] != 8 {
		t.Fatalf("page 1: %v of %d", ids, total)
	}
	last, _, _ := find(t, entryFilter{}, entryOrder{By: "id"}, 99) // past the end: the last page
	if len(last) != pageSize || last[len(last)-1] != 40 {
		t.Errorf("last page: %v", last)
	}
	// The sum is of everything that matches, not of the page.
	_, _, sumOfOne := find(t, entryFilter{Concept: "freelance"}, entryOrder{By: "id"}, 1)
	if sumOfOne <= 0 || sumOfOne >= sum {
		t.Errorf("sum of one = %d, of all = %d", sumOfOne, sum)
	}
}

func TestFilters(t *testing.T) {
	ids, total, _ := find(t, entryFilter{Concept: "GROCER"}, entryOrder{By: "id"}, 1)
	if total == 0 || len(ids) == 0 {
		t.Error("the concept is matched without caring about case")
	}
	if _, total, _ := find(t, entryFilter{Operation: 2}, entryOrder{By: "id"}, 1); total == 0 || total == 40 {
		t.Errorf("by operation: %d", total)
	}
	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, inRange, _ := find(t, entryFilter{From: from, To: to}, entryOrder{By: "id"}, 1)
	_, before, _ := find(t, entryFilter{To: from.AddDate(0, 0, -1)}, entryOrder{By: "id"}, 1)
	if inRange == 0 || before == 0 || inRange+before >= 40 {
		t.Errorf("range %d, before %d", inRange, before)
	}
}

func TestOrderIsStableAndReversible(t *testing.T) {
	asc, _, _ := find(t, entryFilter{}, entryOrder{By: "samount"}, 1)
	desc, _, _ := find(t, entryFilter{}, entryOrder{By: "samount", Desc: true}, 1)
	again, _, _ := find(t, entryFilter{}, entryOrder{By: "samount"}, 1)
	if len(asc) != pageSize || asc[0] == desc[0] {
		t.Errorf("asc %v desc %v", asc, desc)
	}
	for i := range asc {
		if asc[i] != again[i] {
			t.Fatalf("the same order twice gave %v and %v", asc, again)
		}
	}
	// Entries with the same value sort by id, so two pages never repeat an entry.
	seen := map[int]bool{}
	for page := 1; page <= 5; page++ {
		ids, _, _ := find(t, entryFilter{}, entryOrder{By: "category"}, page)
		for _, id := range ids {
			if seen[id] {
				t.Fatalf("entry %d is in two pages", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != 40 {
		t.Errorf("the pages hold %d entries", len(seen))
	}
}

func TestSaveSaysWhoAndWhen(t *testing.T) {
	service := newEntryService()
	e, _ := service.FindOne(1)
	e.Concept = "changed"
	saved, err := service.Save(e)
	if err != nil || saved.CreatedBy != "demo" || !saved.UpdatedAt.After(saved.CreatedAt) || saved.CreatedAt != e.CreatedAt {
		t.Errorf("saved %+v, %v", saved, err)
	}
	if err := service.Delete(1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.FindOne(1); err == nil {
		t.Error("a deleted entry is gone")
	}
}
