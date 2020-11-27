package model

type ChineseRow struct {
	Row       int
	Chinese   string
	Parameter map[string]bool
}

type GroupItem struct {
	Parent string
	Group  string
	URL    string
}

type RouterItem struct {
	Router string
	Method string
}
