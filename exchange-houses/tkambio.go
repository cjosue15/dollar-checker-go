package exchangehouses

import (
	"errors"
	"strconv"

	"github.com/playwright-community/playwright-go"
)

type Tkambio struct {
	Browser *playwright.BrowserContext
}

func NewTkambio(browser *playwright.BrowserContext) *Tkambio {
	return &Tkambio{
		Browser: browser,
	}
}

func (tk *Tkambio) GetExchangeRate() (float64, error) {
	rawPage, err := CreatePage(tk.Browser)

	page := *rawPage

	if _, err = page.Goto("https://tkambio.com/"); err != nil {
		return 0, errors.New("could not goto: %v" + err.Error())
	}

	page.Locator(".exchange-rate.purcharse-content span.price").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})

	entry := page.Locator(".exchange-rate.purcharse-content span.price").First()

	price, err := entry.TextContent()

	if err != nil {
		return 0, errors.New("could not get price: %v" + err.Error())
	}

	floatPrice, err := strconv.ParseFloat(price, 64)

	if err != nil {
		return 0, errors.New("could not parse price: %v" + err.Error())
	}

	return floatPrice, nil
}
