package awss3

import (
	"reflect"
	"testing"
	"time"
)

func TestSortObjects(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	objects := []Object{
		{Name: "old.txt", LastModified: base},
		{Name: "folderA", IsDir: true},
		{Name: "new.txt", LastModified: base.Add(2 * time.Hour)},
		{Name: "folderB", IsDir: true},
		{Name: "mid.txt", LastModified: base.Add(time.Hour)},
	}
	sortObjects(objects)
	got := make([]string, len(objects))
	for i, o := range objects {
		got[i] = o.Name
	}
	want := []string{"folderA", "folderB", "new.txt", "mid.txt", "old.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("排序结果 = %v, 期望 %v", got, want)
	}
}
