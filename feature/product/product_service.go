package product

import (
	"fmt"
	"strings"

	"github.com/deltegui/owl/v3/pagination"

	"owldemo/feature/product/web"
	"owldemo/sys/persistence"
)

const pageSize = 8

type productFilter struct {
	Name     string
	Category string
	Status   string // "active", "disabled" or empty
}

func (f productFilter) match(p productweb.Product) bool {
	return (f.Name == "" || strings.Contains(strings.ToLower(p.Name), strings.ToLower(f.Name))) &&
		(f.Category == "" || p.Category == f.Category) &&
		(f.Status == "" || (f.Status == "active") == p.Active)
}

type productService struct {
	repo persistence.Repository[productweb.Product]
}

func newProductService() productService {
	repo := persistence.NewMemory(
		func(p productweb.Product) int { return p.ID },
		func(p productweb.Product, id int) productweb.Product { p.ID = id; return p },
		seed()...)
	return productService{repo: repo}
}

func seed() []productweb.Product {
	names := []string{"Atlas", "Compass", "Lantern", "Ledger", "Anvil", "Quill", "Beacon", "Sextant", "Bellows", "Chisel"}
	out := make([]productweb.Product, 0, 30)
	for i := 0; i < 30; i++ {
		out = append(out, productweb.Product{
			Name:     fmt.Sprintf("%s %d", names[i%len(names)], i/len(names)+1),
			Category: productweb.Categories[i%len(productweb.Categories)],
			Price:    float64(5+i*3) + 0.99,
			Active:   i%5 != 4,
		})
	}
	return out
}

func (service productService) Find(filter productFilter, p pagination.Pagination) (pagination.List[productweb.Product], error) {
	found, err := service.repo.Find(filter.match, p.CurrentPage, pageSize)
	if err != nil {
		return pagination.List[productweb.Product]{}, err
	}
	return pagination.List[productweb.Product]{
		Items: found.Items,
		Pagination: pagination.Pagination{
			CurrentPage:     found.Page,
			ElementsPerPage: pageSize,
			TotalElements:   found.Total,
		},
	}, nil
}

func (service productService) FindOne(id int) (productweb.Product, error) {
	return service.repo.FindOne(id)
}

func (service productService) Save(p productweb.Product) (productweb.Product, error) {
	return service.repo.Save(p)
}

func (service productService) Delete(id int) error { return service.repo.Delete(id) }

func (service productService) Toggle(id int) (productweb.Product, error) {
	p, err := service.repo.FindOne(id)
	if err != nil {
		return p, err
	}
	p.Active = !p.Active
	return service.repo.Save(p)
}

// Stats is what the chart shows: how many active products there are in each
// category, and what they add up to.
func (service productService) Stats() ([]productweb.CategoryStat, error) {
	all, err := service.repo.Find(func(p productweb.Product) bool { return p.Active }, 1, 1<<20)
	if err != nil {
		return nil, err
	}
	stats := make([]productweb.CategoryStat, 0, len(productweb.Categories))
	for _, category := range productweb.Categories {
		stat := productweb.CategoryStat{Category: category}
		for _, p := range all.Items {
			if p.Category == category {
				stat.Count++
				stat.Value += p.Price
			}
		}
		stats = append(stats, stat)
	}
	return stats, nil
}
