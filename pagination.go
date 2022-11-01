package util

// LimitOffset 用于获取sql分页的limit和offset
func LimitOffset(page, pageSize int64) (limit, offset uint64) {
	limit = uint64(pageSize)
	if page < 1 {
		offset = 1
	}
	offset = uint64((page - 1) * pageSize)

	return
}
