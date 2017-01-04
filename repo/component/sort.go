package component

type catSorter []*Category

func (s catSorter) Len() int           { return len(s) }
func (s catSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s catSorter) Less(i, j int) bool { return s[i].Order < s[j].Order }

type subSorter []*Subcategory

func (s subSorter) Len() int           { return len(s) }
func (s subSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s subSorter) Less(i, j int) bool { return s[i].Order < s[j].Order }

type itemSorter []*Item

func (s itemSorter) Len() int           { return len(s) }
func (s itemSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s itemSorter) Less(i, j int) bool { return s[i].Order < s[j].Order }
