package exchangehouses

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/playwright-community/playwright-go"
)

type Kambista struct {
	Browser *playwright.BrowserContext
}

func NewKambista(browser *playwright.BrowserContext) *Kambista {
	return &Kambista{
		Browser: browser,
	}
}

func (k *Kambista) GetExchangeRate() (float64, error) {
	rawPage, err := CreatePage(k.Browser)

	page := *rawPage

	if _, err = page.Goto("https://www.kambista.com/"); err != nil {
		return 0, errors.New("could not goto: %v" + err.Error())
	}

	page.Locator("#valcompra").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})

	entry := page.Locator("#valcompra").First()

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
