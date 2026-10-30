package contact

import (
	"fmt"
	"strings"

	"github.com/deltegui/owl/v3/pagination"

	"owldemo/feature/contact/web"
	"owldemo/sys/persistence"
)

const pageSize = 8

type contactFilter struct {
	Query  string // matches the name or the email
	Status string // "active", "disabled" or empty
}

func (f contactFilter) match(c contactweb.Contact) bool {
	contains := func(text string) bool {
		return f.Query == "" || strings.Contains(strings.ToLower(text), strings.ToLower(f.Query))
	}
	return (contains(c.Name) || contains(c.Email)) && (f.Status == "" || (f.Status == "active") == c.Active)
}

type contactService struct {
	repo persistence.Repository[contactweb.Contact]
}

func newContactService() contactService {
	repo := persistence.NewMemory(
		func(c contactweb.Contact) int { return c.ID },
		func(c contactweb.Contact, id int) contactweb.Contact { c.ID = id; return c },
		seed()...)
	return contactService{repo: repo}
}

func seed() []contactweb.Contact {
	names := []string{"Ada Lovelace", "Alan Turing", "Grace Hopper", "Edsger Dijkstra", "Barbara Liskov", "Ken Thompson", "Rob Pike", "Margaret Hamilton"}
	out := make([]contactweb.Contact, 0, 20)
	for i := 0; i < 20; i++ {
		out = append(out, contactweb.Contact{
			Name:   fmt.Sprintf("%s %d", names[i%len(names)], i/len(names)+1),
			Email:  fmt.Sprintf("contact%d@example.com", i+1),
			Phone:  fmt.Sprintf("+34 600 000 %03d", i),
			Active: i%5 != 4,
		})
	}
	return out
}

func (service contactService) Find(filter contactFilter, p pagination.Pagination) (pagination.List[contactweb.Contact], error) {
	found, err := service.repo.Find(filter.match, p.CurrentPage, pageSize)
	if err != nil {
		return pagination.List[contactweb.Contact]{}, err
	}
	return pagination.List[contactweb.Contact]{
		Items: found.Items,
		Pagination: pagination.Pagination{
			CurrentPage:     found.Page,
			ElementsPerPage: pageSize,
			TotalElements:   found.Total,
		},
	}, nil
}

func (service contactService) Save(c contactweb.Contact) (contactweb.Contact, error) {
	return service.repo.Save(c)
}

func (service contactService) Delete(id int) error { return service.repo.Delete(id) }

func (service contactService) Toggle(id int) (contactweb.Contact, error) {
	c, err := service.repo.FindOne(id)
	if err != nil {
		return c, err
	}
	c.Active = !c.Active
	return service.repo.Save(c)
}
