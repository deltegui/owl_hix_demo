package menu

type menuService struct{}

// items are the links of the side menu. The key is the text in common/menu.
func (menuService) items() []struct{ Href, Key string } {
	return []struct{ Href, Key string }{
		{"/products", "Menu.Products"},
		{"/entries", "Menu.Entries"},
		{"/contacts", "Menu.Contacts"},
		{"/forms/product", "Menu.NewProduct"},
		{"/forms/contact", "Menu.NewContact"},
		{"/about", "Menu.About"},
		{"/bulma", "Menu.Bulma"},
		{"/coat", "Menu.Coat"},
		{"/bs", "Menu.Bs"},
	}
}
