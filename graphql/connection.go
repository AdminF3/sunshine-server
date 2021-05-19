package graphql

import (
	"encoding/base64"
	"strconv"
	"strings"
)

// encodeCursor takes index and converts it to a GraphQL cursor.
func encodeCursor(index int) string {
	return base64.StdEncoding.EncodeToString([]byte("cursor:" + strconv.Itoa(index)))
}

// decodeCursor takes an after cursor and converts to index.
func decodeCursor(cursor *string) int {
	if cursor == nil {
		return -1
	}

	dec, err := base64.StdEncoding.DecodeString(*cursor)
	if err != nil {
		return -1
	}

	parts := strings.Split(string(dec), ":")
	if len(parts) < 2 {
		return -1
	}

	index, err := strconv.Atoi(parts[1])
	if err != nil || index < -1 {
		return -1
	}

	return index
}

// calcBounds calculates offset and limit out of first, after, last and before
// argument. Users should pass either first with after OR last with before.
//
// However, even though the spec discourages passing both, it allows it. In
// this case this function returns the intersection of both ranges. If the
// intersection is an empty set last and before are ignored.
func calcBounds(first, after, last, before, total int) (offset, limit int) {
	if after >= 0 {
		offset = after + 1
	}
	if first > 0 {
		limit = first
	}

	// last should change the limit only when after is not passed or when
	// applying last narrows down the selection.
	if last > 0 {
		if limit == 0 || limit > last {
			limit = last
		}

		if before < 1 && total >= last {
			offset = total - last
		}
	}

	// change offset only if before makes the selection narrower.
	if before > 0 && before-last > offset {
		if limit == 0 {
			limit = before
		} else {
			offset = before - last
		}
	}

	return offset, limit
}

// pageInfo creates PageInfo value out of given bounds.
// Note that pageInfo doesn't populate StartCursor and EndCursor.
func pageInfo(offset, limit, length int) *PageInfo {
	return &PageInfo{
		HasNextPage:     0 < offset+limit && offset+limit < length,
		HasPreviousPage: offset > 0,
	}
}
