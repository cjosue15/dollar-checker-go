package exchangehouses

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/playwright-community/playwright-go"
)

type Rextie struct {
	Browser *playwright.BrowserContext
}

func NewRextie(browser *playwright.BrowserContext) *Rextie {
	return &Rextie{
		Browser: browser,
	}
}

func (r *Rextie) GetExchangeRate() (float64, error) {
	rawPage, err := CreatePage(r.Browser)

	page := *rawPage

	if _, err = page.Goto("https://www.rextie.com/"); err != nil {
		return 0, errors.New("could not goto: %v" + err.Error())
	}

	page.Locator(".app-gql-exchange-rate > div > div > div > div").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})

	entry := page.Locator("app-gql-exchange-rate > div > div > div > div div:nth-child(2)").First()

	rawPrice, err := entry.TextContent()

	if err != nil {
		return 0, errors.New("could not get price: %v" + err.Error())
	}

	re := regexp.MustCompile(`\d+\.\d+`)
	match := re.FindString(rawPrice)
	floatPrice, err := strconv.ParseFloat(match, 64)

	if err != nil {
		return 0, errors.New("could not parse price: %v" + err.Error())
	}

	return floatPrice, nil
}
