package v1

type Action struct {
	Name  string      `json:"name" binding:"required,min=1,max=64,excludesall=\u002F\u005C"`
	Value interface{} `json:"value" binding:"required"`
}
