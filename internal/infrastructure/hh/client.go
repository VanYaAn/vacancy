package hh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"mcpPrep/internal/domain"
)

const (
	baseURL   = "https://api.hh.ru"
	userAgent = "mcpPrep/1.0"
)

type Client struct {
	http      *http.Client
	token     string
	rateLimit time.Duration
	lastReq   time.Time
}

func NewClient(token string) *Client {
	return &Client{
		http:      &http.Client{Timeout: 15 * time.Second},
		token:     token,
		rateLimit: 300 * time.Millisecond,
	}
}

// SearchVacancies — поиск вакансий по параметрам, возвращает одну страницу.
func (c *Client) SearchVacancies(params domain.SearchParams) (*domain.SearchResult, error) {
	body, err := c.do("GET", "/vacancies", paramsToQuery(params))
	if err != nil {
		return nil, fmt.Errorf("search vacancies: %w", err)
	}

	var raw apiSearchResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	return mapSearchResult(raw), nil
}

// GetVacancy — полные данные одной вакансии.
func (c *Client) GetVacancy(id string) (*domain.VacancyDetail, error) {
	body, err := c.do("GET", "/vacancies/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("get vacancy %s: %w", id, err)
	}

	var raw apiVacancyDetail
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse vacancy %s: %w", id, err)
	}

	return mapVacancyDetail(raw), nil
}

// SearchAll — постранично собирает все вакансии (до maxPages страниц).
func (c *Client) SearchAll(params domain.SearchParams, maxPages int) ([]domain.Vacancy, error) {
	var all []domain.Vacancy
	params.Page = 0

	for {
		result, err := c.SearchVacancies(params)
		if err != nil {
			return nil, err
		}
		all = append(all, result.Items...)

		if result.Page+1 >= result.Pages || (maxPages > 0 && result.Page+1 >= maxPages) {
			break
		}
		params.Page++
	}
	return all, nil
}

// ─── internal ─────────────────────────────────────────────────────────────────

func (c *Client) do(method, path string, query url.Values) ([]byte, error) {
	if since := time.Since(c.lastReq); since < c.rateLimit {
		time.Sleep(c.rateLimit - since)
	}
	c.lastReq = time.Now()

	u := baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("HH-User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusTooManyRequests:
		return nil, ErrRateLimit
	default:
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
}

func paramsToQuery(p domain.SearchParams) url.Values {
	q := url.Values{}
	if p.Text != "" {
		q.Set("text", p.Text)
	}
	if p.Area != "" {
		q.Set("area", p.Area)
	}
	if p.Experience != "" {
		q.Set("experience", p.Experience)
	}
	if p.WorkFormat != "" {
		q.Set("work_format", p.WorkFormat)
	}
	if p.Salary > 0 {
		q.Set("salary", strconv.Itoa(p.Salary))
		if p.Currency != "" {
			q.Set("currency", p.Currency)
		}
	}
	if p.OnlyWithSalary {
		q.Set("label", "with_salary")
	}
	perPage := p.PerPage
	if perPage == 0 {
		perPage = 20
	}
	q.Set("per_page", strconv.Itoa(perPage))
	q.Set("page", strconv.Itoa(p.Page))
	return q
}
