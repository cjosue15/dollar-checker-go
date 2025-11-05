package exchangehouses

import (
	"errors"
	"github.com/playwright-community/playwright-go"
)

func CreatePage(browser *playwright.BrowserContext) (*playwright.Page, error) {
	page, err := (*browser).NewPage()
	if err != nil {
		return nil, errors.New("could not create page")
	}

	return &page, nil
}
