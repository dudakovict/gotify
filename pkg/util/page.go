package util

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Page represents the requested page and rows per page.
type Page struct {
	Number      int
	RowsPerPage int
}

// Parse parses the request for the page and rows query string. The
// defaults are provided as well.
func Parse(c *fiber.Ctx) (Page, error) {
	var number int
	if page := c.Query("page", "1"); page != "" {
		var err error
		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, err
		}
	}

	var rowsPerPage int
	if rows := c.Query("rows", "10"); rows != "" {
		var err error
		rowsPerPage, err = strconv.Atoi(rows)
		if err != nil {
			return Page{}, err
		}
	}

	p := Page{
		Number:      number,
		RowsPerPage: rowsPerPage,
	}

	return p, nil
}
