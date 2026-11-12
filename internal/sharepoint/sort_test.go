package sharepoint

import (
	"reflect"
	"testing"
	"time"
)

func TestSortItems(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	items := []Item{
		{Name: "old.txt", LastModified: base},
		{Name: "folderOld", IsFolder: true, LastModified: base},
		{Name: "new.txt", LastModified: base.Add(2 * time.Hour)},
		{Name: "folderNew", IsFolder: true, LastModified: base.Add(3 * time.Hour)},
		{Name: "mid.txt", LastModified: base.Add(time.Hour)},
	}
	sortItems(items)
	got := make([]string, len(items))
	for i, it := range items {
		got[i] = it.Name
	}
	want := []string{"folderNew", "folderOld", "new.txt", "mid.txt", "old.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("排序结果 = %v, 期望 %v", got, want)
	}
}
