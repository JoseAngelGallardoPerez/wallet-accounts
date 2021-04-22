package response

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/gin-gonic/gin"
)

type Error struct {
	Title   string       `json:"title"`
	Details string       `json:"details"`
	Status  int          `json:"status"`
	Code    *string      `json:"code"`
	Source  *ErrorSource `json:"source"`
}

type ErrorSource struct {
	Pointer string `json:"pointer"`
}

type List struct {
	HasMore bool        `json:"hasMore"`
	Items   interface{} `json:"items"`
}

type Response struct {
	Links      interface{} `json:"links,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Messages   []string    `json:"messages,omitempty"`
	Errors     []*Error    `json:"errors,omitempty"`
}

type Links struct {
	Self  *string `json:"self"`
	Next  *string `json:"next"`
	Prev  *string `json:"prev"`
	First *string `json:"first"`
	Last  *string `json:"last"`
}

type Pagination struct {
	TotalRecord uint64 `json:"totalRecord"`
	TotalPage   uint64 `json:"totalPage"`
	Offset      uint64 `json:"offset"`
	Limit       uint64 `json:"limit"`
	PageNumber  uint64 `json:"pageNumber"`
	CurrentPage uint64 `json:"currentPage"`
}

func New() *Response {
	return new(Response)
}

func NewList(items interface{}) *List {
	result := &List{}
	result.Items = items

	return result
}

func NewWithList(items interface{}) *Response {
	list := NewList(items)

	res := New()
	res.SetData(list.Items)

	return res
}

func NewWithListAndPageLinks(items interface{}, total uint64,
	path string, listParams *list_params.ListParams,
) *Response {
	list := NewList(items)
	res := New()
	res.SetData(list.Items)
	unescapedPath, err := url.PathUnescape(path)

	if err != nil {
		log.Println(err)
	}

	links := buildPageLinks(unescapedPath, (uint64)(listParams.Pagination.PageNumber),
		(uint64)(listParams.Pagination.PageSize), total)
	res.SetLinks(links)

	return res
}

func NewWithListAndLinksAndPagination(items interface{}, total uint64,
	path string, listParams *list_params.ListParams,
) *Response {
	res := NewWithListAndPageLinks(items, total, path, listParams)

	pagination := paginate((uint64)(listParams.Pagination.PageNumber),
		(uint64)(listParams.Pagination.PageSize), total)
	res.SetPagination(pagination)

	return res
}

func NewWithListAndPagination(items interface{}, total uint64, listParams *list_params.ListParams) *Response {
	res := NewWithList(items)

	pagination := paginate((uint64)(listParams.Pagination.PageNumber),
		(uint64)(listParams.Pagination.PageSize), total)
	res.SetPagination(pagination)

	return res
}

func NewWithError(
	title string,
	details string,
	status int,
	code *string,
) *Response {
	return New().AddError(title, details, status, code)
}

func NewCommonError(status int) *Response {
	return New().AddError(
		http.StatusText(status), http.StatusText(status), status, nil)
}

func (r *Response) AddError(
	title string,
	details string,
	status int,
	code *string,
) *Response {
	e := &Error{
		Title:   title,
		Details: details,
		Status:  status,
		Code:    code,
	}
	r.Errors = append(r.Errors, e)
	return r
}

func (r *Response) AddMessage(message string) *Response {
	r.Messages = append(r.Messages, message)
	return r
}

func (r *Response) SetData(data interface{}) *Response {
	r.Data = data
	return r
}

func (r *Response) SetLinks(links interface{}) *Response {
	r.Links = links
	return r
}

func (r *Response) SetPagination(pagination *Pagination) *Response {
	r.Pagination = pagination
	return r
}

func (r *Response) buildLinks(c *gin.Context, total int64) Links {
	limit, err := strconv.ParseInt(c.Request.URL.Query().Get("limit"), 10, 32)
	if nil != err {
		limit = 15
	}

	selfLink := c.Request.URL.String()
	links := Links{
		Self:  &selfLink,
		Next:  r.getNextUrl(c, total, limit),
		Prev:  r.getPrevUrl(c, total, limit),
		First: r.getFirstUrl(c, total),
		Last:  r.getLastUrl(c, total, limit),
	}

	return links
}

func (r *Response) getNextUrl(c *gin.Context, total int64, limit int64) *string {
	offset, _ := strconv.ParseInt(c.Request.URL.Query().Get("offset"), 10, 32)

	if offset > total || offset+limit > total {
		return nil
	}

	offsetNext := limit + offset

	url := *c.Request.URL

	values := url.Query()
	values.Set("offset", strconv.Itoa(int(offsetNext)))
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}

func (r *Response) getPrevUrl(c *gin.Context, total int64, limit int64) *string {
	offset, _ := strconv.ParseInt(c.Request.URL.Query().Get("offset"), 10, 32)

	if offset == 0 {
		return nil
	}

	url := *c.Request.URL

	offsetNext := offset - limit
	values := url.Query()
	values.Set("offset", strconv.Itoa(int(offsetNext)))
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}

func (r *Response) getFirstUrl(c *gin.Context, total int64) *string {
	if total == 0 {
		return nil
	}

	url := *c.Request.URL

	values := url.Query()
	values.Set("offset", "0")
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}

func (r *Response) getLastUrl(c *gin.Context, total int64, limit int64) *string {
	if total == 0 {
		return nil
	}

	pages := int64(math.Round(float64(total / limit)))
	lastOffset := limit * pages

	url := *c.Request.URL

	values := url.Query()
	values.Set("offset", strconv.Itoa(int(lastOffset)))
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}

func buildPageLinks(path string, number, size, total uint64) Links {
	links := Links{
		First: buildFirstPath(path, size),
		Last:  buildLastPath(path, size, total),
		Next:  buildNextPath(path, number, size, total),
		Prev:  buildPrevPath(path, number, size, total),
		Self:  buildSelfPath(path, number, size),
	}

	return links
}

func paginate(number, size, total uint64) *Pagination {
	offset := size * (number - 1)
	totalPage := uint64(math.Ceil(float64(total / size)))

	return &Pagination{
		Limit:       size,
		PageNumber:  number,
		CurrentPage: number,
		Offset:      offset,
		TotalRecord: total,
		TotalPage:   totalPage,
	}
}

func buildFirstPath(path string, size uint64) *string {
	if size == 0 {
		return &path
	}
	return changePageNumber(path, 1)
}

func buildLastPath(path string, size, total uint64) *string {
	if size < 1 {
		return nil
	}
	return changePageNumber(path, getLastPageNumber(size, total))
}

func buildNextPath(path string, number, size, total uint64) *string {
	if size == 0 {
		return &path
	}

	lastPageNumber := getLastPageNumber(size, total)
	if number >= lastPageNumber {
		return nil
	}
	return changePageNumber(path, number+1)
}

func buildSelfPath(path string, number, size uint64) *string {
	if size == 0 {
		return &path
	}
	return changePageNumber(path, number)
}

func buildPrevPath(path string, number, size, total uint64) *string {
	if number <= 1 {
		return nil
	}

	lastPageNumber := getLastPageNumber(size, total)
	if number > lastPageNumber {
		return changePageNumber(path, lastPageNumber)
	}

	return changePageNumber(path, number-1)
}

func changePageNumber(input string, to uint64) *string {
	re := regexp.MustCompile(`(page\[number\])(=\d+)`)
	strToParse := input
	if !re.MatchString(input) {
		strToParse = addParameterToPath(input, "page[number]=0")

	}
	result := re.ReplaceAllString(strToParse, fmt.Sprintf(`$1=%v`, to))
	return &result
}

func addParameterToPath(path, parameter string) string {
	var delimeter string
	if strings.Contains(path, "?") {
		delimeter = "&"
	} else {
		delimeter = "?"
	}
	return path + delimeter + parameter
}

func getLastPageNumber(size, total uint64) uint64 {
	count := float64(total) / float64(size)
	return uint64(math.Ceil(count))
}
