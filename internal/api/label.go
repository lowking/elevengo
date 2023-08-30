package api

import (
	"github.com/deadblue/elevengo/internal/api/base"
)

const (
	LabelListLimit = 30

	LabelColorBlank  = "#000000"
	LabelColorRed    = "#FF4B30"
	LabelColorOrange = "#F78C26"
	LabelColorYellow = "#FFC032"
	LabelColorGreen  = "#43BA80"
	LabelColorBlue   = "#2670FC"
	LabelColorPurple = "#8B69FE"
	LabelColorGray   = "#CCCCCC"
)

type LabelInfo struct {
	Id         string         `json:"id"`
	Name       string         `json:"name"`
	Color      string         `json:"color"`
	Sort       base.IntNumber `json:"sort"`
	CreateTime int64          `json:"create_time"`
	UpdateTime int64          `json:"update_time"`
}

type LabelListResult struct {
	Total int          `json:"total"`
	List  []*LabelInfo `json:"list"`
	Sort  string       `json:"sort"`
	Order string       `json:"order"`
}

type LabelListSpec struct {
	base.JsonApiSpec[LabelListResult, base.StandardResp]
}

func (s *LabelListSpec) Init(offset int) *LabelListSpec {
	s.JsonApiSpec.Init("https://webapi.115.com/label/list")
	s.QuerySetInt("offset", offset)
	s.QuerySetInt("limit", LabelListLimit)
	s.QuerySet("sort", "create_time")
	s.QuerySet("order", "asc")
	return s
}

type LabelSearchSpec struct {
	base.JsonApiSpec[LabelListResult, base.StandardResp]
}

func (s *LabelSearchSpec) Init(keyword string, offset int) *LabelSearchSpec {
	s.JsonApiSpec.Init("https://webapi.115.com/label/list")
	s.QuerySet("keyword", keyword)
	s.QuerySetInt("offset", offset)
	s.QuerySetInt("limit", LabelListLimit)
	return s
}

type LabelCreateResult []*LabelInfo

type LabelCreateSpec struct {
	base.JsonApiSpec[LabelCreateResult, base.StandardResp]
}

func (s *LabelCreateSpec) Init(name, color string) *LabelCreateSpec {
	s.JsonApiSpec.Init("https://webapi.115.com/label/add_multi")
	s.FormSet("name[]", name+"\x07"+color)
	return s
}

type LabelEditSpec struct {
	base.JsonApiSpec[base.VoidResult, base.BasicResp]
}

func (s *LabelEditSpec) Init(labelId, name, color string) *LabelEditSpec {
	s.JsonApiSpec.Init("https://webapi.115.com/label/edit")
	s.FormSetAll(map[string]string{
		"id":    labelId,
		"name":  name,
		"color": color,
	})
	return s
}

type LabelDeleteSpec struct {
	base.JsonApiSpec[base.VoidResult, base.BasicResp]
}

func (s *LabelDeleteSpec) Init(labelId string) *LabelDeleteSpec {
	s.JsonApiSpec.Init("https://webapi.115.com/label/delete")
	s.FormSet("id", labelId)
	return s
}

type LabelSetOrderSpec struct {
	base.JsonApiSpec[base.VoidResult, base.BasicResp]
}

func (s *LabelSetOrderSpec) Init(labelId string, order string, asc bool) *LabelSetOrderSpec {
	s.JsonApiSpec.Init("https://webapi.115.com/files/order")
	s.FormSetAll(map[string]string{
		"module":     "label_search",
		"file_id":    labelId,
		"fc_mix":     "0",
		"user_order": order,
	})
	if asc {
		s.FormSet("user_asc", "1")
	} else {
		s.FormSet("user_asc", "0")
	}
	return s
}
