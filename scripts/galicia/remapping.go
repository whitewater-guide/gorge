package galicia

import "fmt"

// remapping of new station id to old station ids
var remapping = map[int]string{
	140515: "30546",
	140470: "30568",
	140530: "30552",
	140508: "30542",
	141720: "30438",
	140365: "30578",
	141230: "30464",
	141560: "30443",
	140165: "30588",
	141780: "30433",
	141670: "30441",
	140510: "30544",
	140540: "30554",
	140150: "30585",
}

func gaugeCode(id int) string {
	if old, ok := remapping[id]; ok {
		return old
	}
	return fmt.Sprint(id)
}
