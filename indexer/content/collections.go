package content

import (
	"sort"
	"time"
)

type Collection []ContentFile

type Collections = map[string]Collection

func (collection Collection) SortByDate() Collection {
	sort.Slice(collection, func(i, j int) bool {
		date1, err := time.Parse("2006-01-02", collection[i].MetaData["date"].(string))
		if err != nil {
			return false
		}

		date2, err := time.Parse("2006-01-02", collection[j].MetaData["date"].(string))
		if err != nil {
			return false
		}

		return date1.After(date2)
	})

	return collection
}

func (collection Collection) Slice(start int, end int) Collection {
	if end > len(collection) {
		end = len(collection)
	}
	return collection[start:end]
}
