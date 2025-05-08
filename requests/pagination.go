package requests

type Pagination struct {
	Limit  int `json:"limit" query:"limit"`
	Offset int `json:"offset" query:"offset"`
}

func GetPagination(pagination Pagination) Pagination {
	if pagination.Limit == 0 {
		pagination.Limit = 5
	}
	if pagination.Limit > 15 {
		pagination.Limit = 15
	}
	return pagination
}
