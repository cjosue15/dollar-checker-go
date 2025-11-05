package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	exchangehouses "github.com/cjosue15/dollar-checker-cli/exchange-houses"
	"github.com/playwright-community/playwright-go"
)

type ExchangeRate struct {
	Name  string
	Price float64
	Error error
}

func main() {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Obtaining exchange rates...\n"
	s.Start()

	bw, pw, err := initPlaywright()

	if err != nil {
		log.Fatalf("could not initialize playwright: %v", err)
	}

	defer closePlaywright(bw, pw)

	rates := fetchRates(bw)

	s.Stop()

	// filter options
	var options []huh.Option[string]
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

	for _, rate := range rates {
		if rate.Error == nil {
			label := fmt.Sprintf("%s: S/ %.3f", rate.Name, rate.Price)
			fmt.Println(successStyle.Render("✓ " + label))
			options = append(options, huh.NewOption(label, rate.Name))
		} else {
			errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
			fmt.Println(errorStyle.Render(fmt.Sprintf("✗ %s: %v", rate.Name, rate.Error)))
		}
	}

	if len(options) == 0 {
		fmt.Println("\n❌ No se pudo obtener ningún tipo de cambio")
		return
	}

	// show form
	var selected string
	var amount string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Selecciona una casa de cambio:").
				Options(options...).
				Value(&selected),

			huh.NewInput().
				Title("Monto en dólares a cambiar:").
				Validate(
					func(input string) error {
						if _, err := strconv.ParseFloat(input, 64); err != nil {
							return errors.New("please enter a valid number")
						}
						return nil
					},
				).
				Value(&amount),
		),
	)

	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	amountFloat, ok := strconv.ParseFloat(amount, 64)
	if ok != nil {
		fmt.Println("❌ Monto inválido")

	}

	fmt.Printf("\n✅ Conversion: %v\n", getAmountInSoles(rates, selected, amountFloat))
}

func getAmountInSoles(rates []ExchangeRate, house string, amount float64) float64 {
	var selectedRate *ExchangeRate
	for _, rate := range rates {
		if rate.Name == house {
			selectedRate = &rate
			break
		}
	}
	return amount * selectedRate.Price
}

func initPlaywright() (*playwright.Browser, *playwright.Playwright, error) {
	err := playwright.Install()

	if err != nil {
		return nil, nil, errors.New("could not install playwright")
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, errors.New("could not start playwright")
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--no-sandbox",
		},
	})
	if err != nil {
		return nil, nil, errors.New("could not launch browser")
	}

	return &browser, pw, nil
}

func closePlaywright(bw *playwright.Browser, pw *playwright.Playwright) {
	if bw != nil {
		(*bw).Close()
	}

	if pw != nil {
		pw.Stop()
	}
}

func fetchRates(bw *playwright.Browser) []ExchangeRate {
	var wg sync.WaitGroup
	var mu sync.Mutex
	rates := []ExchangeRate{}
	browser := *bw
	contexts := make([]playwright.BrowserContext, 3)

	for i := range contexts {
		contexts[i], _ = browser.NewContext()
		defer contexts[i].Close()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		tk := exchangehouses.NewTkambio(&contexts[0])
		price, err := tk.GetExchangeRate()
		mu.Lock()
		rates = append(rates, ExchangeRate{Name: "Tkambio", Price: price, Error: err})
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r := exchangehouses.NewRextie(&contexts[1])
		priceRextie, err := r.GetExchangeRate()
		mu.Lock()
		rates = append(rates, ExchangeRate{Name: "Rextie", Price: priceRextie, Error: err})
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		k := exchangehouses.NewKambista(&contexts[2])
		priceKambista, err := k.GetExchangeRate()
		mu.Lock()
		rates = append(rates, ExchangeRate{Name: "Kambista", Price: priceKambista, Error: err})
		mu.Unlock()
	}()

	wg.Wait()
	return rates
}
