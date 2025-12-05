package types

type SetPayloadQuery struct {
	QueryUser        string `json:"search_u"` // user search input
	QueryRole        string `json:"search_r"` // role search input
	QueryMember      string `json:"search_m"` // member search input
	QueryBook        string `json:"search_b"` // book search input
	QueryCirculation string `json:"search_c"` // circulation search input
}
