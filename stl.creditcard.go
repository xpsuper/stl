package stl

import (
	"errors"
	"strconv"
	"time"
)

// CreditCard holds generic information about the credit card
type CreditCard struct {
	Number, Cvv, Month, Year string
	Company                  Company
}

// Company holds a short and long names of who has issued the credit card
type Company struct {
	Short, Long string
}

type digits [6]int

// At returns the digits from the start to the given length
func (d *digits) At(i int) int {
	return d[i-1]
}

// LastFour returns the last four digits of the credit card's number
func (c *CreditCard) LastFour() (string, error) {
	if len(c.Number) < 4 {
		return "", errors.New("Credit card number is not long enough")
	}

	return c.Number[len(c.Number)-4 : len(c.Number)], nil
}

// LastFourDigits as an alias for LastFour
func (c *CreditCard) LastFourDigits() (string, error) {
	return c.LastFour()
}

// Wipe returns the credit card with false/nullified/generic information
func (c *CreditCard) Wipe() {
	c.Cvv, c.Number, c.Month, c.Year = "0000", "0000000000000000", "01", "1970"
}

// Validate returns nil or an error describing why the credit card didn't validate
// this method checks for expiration date, CCV/CVV and the credit card's numbers.
// For allowing test cards to go through, simply pass true (bool) as the first argument
func (c *CreditCard) Validate(allowTestNumbers ...bool) error {
	err := c.ValidateExpiration()
	if err != nil {
		return err
	}

	err = c.ValidateCVV()
	if err != nil {
		return err
	}

	switch c.Number {
	case "4242424242424242",
		"4012888888881881",
		"4000056655665556",
		"5555555555554444",
		"5200828282828210",
		"5105105105105100",
		"378282246310005",
		"371449635398431",
		"6011111111111117",
		"6011000990139424",
		"30569309025904",
		"38520000023237",
		"3530111333300000",
		"3566002020360505",
		"4111111111111111",
		"4916909992637469",
		"4000111111111115",
		"2223000048400011",
		"6035227716427021":
		if len(allowTestNumbers) > 0 && allowTestNumbers[0] {
			return nil
		}

		return errors.New("Test numbers are not allowed")
	}

	valid := c.ValidateNumber()

	if !valid {
		return errors.New("Invalid credit card number")
	}

	return nil
}

func (c *CreditCard) ValidateExpiration() error {
	var year, month int
	var err error
	timeNow := time.Now()

	if len(c.Year) < 3 {
		year, err = strconv.Atoi(strconv.Itoa(timeNow.UTC().Year())[:2] + c.Year)
		if err != nil {
			return errors.New("Invalid year")
		}
	} else {
		year, err = strconv.Atoi(c.Year)
		if err != nil {
			return errors.New("Invalid year")
		}
	}

	month, err = strconv.Atoi(c.Month)
	if err != nil {
		return errors.New("Invalid month")
	}

	if month < 1 || 12 < month {
		return errors.New("Invalid month")
	}

	if year < timeNow.UTC().Year() {
		return errors.New("Credit card has expired")
	}

	if year == timeNow.UTC().Year() && month < int(timeNow.UTC().Month()) {
		return errors.New("Credit card has expired")
	}

	return nil
}

func (c *CreditCard) ValidateCVV() error {
	if len(c.Cvv) < 3 || len(c.Cvv) > 4 {
		return errors.New("Invalid CVV")
	}

	return nil
}

// Method returns an error from MethodValidate() or returns the
// credit card with its company / issuer attached to it
func (c *CreditCard) Method() error {
	company, err := c.MethodValidate()

	if err != nil {
		return err
	}

	c.Company = company
	return nil
}

// MethodValidate adds/checks/verifies the credit card's company / issuer
func (c *CreditCard) MethodValidate() (Company, error) {
	var err error
	ccLen := len(c.Number)
	ccDigits := digits{}

	for i := 0; i < 6; i++ {
		if i < ccLen {
			ccDigits[i], err = strconv.Atoi(c.Number[:i+1])
			if err != nil {
				return Company{"", ""}, errors.New("Unknown credit card method")
			}
		}
	}

	switch {
	case isAmex(ccDigits):
		return Company{"amex", "American Express"}, nil
	case isBankCard(ccDigits):
		return Company{"bankcard", "Bankcard"}, nil
	case isCabal(ccDigits):
		return Company{"cabal", "Cabal"}, nil
	case isUnionPay(ccDigits):
		return Company{"china unionpay", "China UnionPay"}, nil
	case isDinersClubCarteBlanche(ccDigits, ccLen):
		return Company{"diners club carte blanche", "Diners Club Carte Blanche"}, nil
	case isDinersClubEnroute(ccDigits):
		return Company{"diners club enroute", "Diners Club enRoute"}, nil
	case isDinersClubInternational(ccDigits, ccLen):
		return Company{"diners club international", "Diners Club International"}, nil
	case isDiscover(ccDigits):
		return Company{"discover", "Discover"}, nil
	// Elo must be checked before interpayment
	case isElo(ccDigits):
		return Company{"elo", "Elo"}, nil
	case isHiperCard(ccDigits):
		return Company{"hipercard", "Hipercard"}, nil
	case isInterPayment(ccDigits, ccLen):
		return Company{"interpayment", "InterPayment"}, nil
	case isInstaPayment(ccDigits, ccLen):
		return Company{"instapayment", "InstaPayment"}, nil
	case isJCB(ccDigits):
		return Company{"jcb", "JCB"}, nil
	case isNaranJa(ccDigits):
		return Company{"naranja", "Naranja"}, nil
	case isMaestro(c, ccDigits):
		return Company{"maestro", "Maestro"}, nil
	case isDankort(ccDigits):
		return Company{"dankort", "Dankort"}, nil
	case isMasterCard(ccDigits):
		return Company{"mastercard", "MasterBankCard"}, nil
	case isVisaElectron(ccDigits):
		return Company{"visa electron", "Visa Electron"}, nil
	case isVisa(ccDigits):
		return Company{"visa", "Visa"}, nil
	case isAura(ccDigits):
		return Company{"aura", "Aura"}, nil
	default:
		return Company{"", ""}, errors.New("Unknown credit card method")
	}
}

// Luhn algorithm
// http://en.wikipedia.org/wiki/Luhn_algorithm

// ValidateNumber will check the credit card's number against the Luhn algorithm
func (c *CreditCard) ValidateNumber() bool {
	var sum int
	var alternate bool

	numberLen := len(c.Number)

	if numberLen < 13 || numberLen > 19 {
		return false
	}

	for i := numberLen - 1; i > -1; i-- {
		mod, _ := strconv.Atoi(string(c.Number[i]))
		if alternate {
			mod *= 2
			if mod > 9 {
				mod = (mod % 10) + 1
			}
		}

		alternate = !alternate

		sum += mod
	}

	return sum%10 == 0
}

func matchesValue(number int, numbers []int) bool {
	for _, v := range numbers {
		if v == number {
			return true
		}
	}
	return false
}

func isInBetween(n, min, max int) bool {
	return n >= min && n <= max
}

func isAmex(ccDigits digits) bool {
	return matchesValue(ccDigits.At(2), []int{34, 37})
}

func isBankCard(ccDigits digits) bool {
	return ccDigits.At(4) == 5610 || isInBetween(ccDigits.At(6), 560221, 560225)
}

func isCabal(ccDigits digits) bool {
	atSix := ccDigits.At(6)

	return matchesValue(atSix, []int{604400, 627170, 603522, 589657}) ||
		isInBetween(atSix, 604201, 604219) ||
		isInBetween(atSix, 604300, 604399)
}

func isUnionPay(ccDigits digits) bool {
	return matchesValue(ccDigits.At(2), []int{62, 81})
}

func isDinersClubCarteBlanche(ccDigits digits, ccLen int) bool {
	return isInBetween(ccDigits.At(3), 300, 305) && ccLen == 14
}

func isDinersClubEnroute(ccDigits digits) bool {
	return matchesValue(ccDigits.At(4), []int{2014, 2149})
}

func isDinersClubInternational(ccDigits digits, ccLen int) bool {
	checkThree := isInBetween(ccDigits.At(3), 300, 305) || ccDigits.At(3) == 309
	checkTwo := matchesValue(ccDigits.At(2), []int{36, 38, 39})

	return (checkThree || checkTwo) && ccLen <= 14
}

func isDiscover(ccDigits digits) bool {
	return ccDigits.At(4) == 6011 ||
		isInBetween(ccDigits.At(6), 622126, 622925) ||
		isInBetween(ccDigits.At(3), 644, 649) ||
		ccDigits.At(2) == 65
}

func isElo(ccDigits digits) bool {
	atFour := ccDigits.At(4)
	atSix := ccDigits.At(6)

	return matchesValue(atFour, []int{4011, 4576}) ||
		matchesValue(atSix, []int{431274, 438935, 451416, 457393, 457631, 457632, 504175, 627780, 636297, 636368, 636369}) ||
		isInBetween(atSix, 506699, 506778) ||
		isInBetween(atSix, 509000, 509999) ||
		isInBetween(atSix, 650031, 650051) ||
		isInBetween(atSix, 650035, 650033) ||
		isInBetween(atSix, 650405, 650439) ||
		isInBetween(atSix, 650485, 650538) ||
		isInBetween(atSix, 650541, 650598) ||
		isInBetween(atSix, 650700, 650718) ||
		isInBetween(atSix, 650720, 650727) ||
		isInBetween(atSix, 650901, 650920) ||
		isInBetween(atSix, 651652, 651679) ||
		isInBetween(atSix, 655000, 655019) ||
		isInBetween(atSix, 655021, 655021)
}

func isHiperCard(ccDigits digits) bool {
	return matchesValue(ccDigits.At(6), []int{606282, 637095, 637568, 637599, 637609, 637612})
}

func isInterPayment(ccDigits digits, ccLen int) bool {
	return ccDigits.At(3) == 636 && isInBetween(ccLen, 16, 19)
}

func isInstaPayment(ccDigits digits, ccLen int) bool {
	return isInBetween(ccDigits.At(3), 637, 639) && ccLen == 16
}

func isJCB(ccDigits digits) bool {
	return isInBetween(ccDigits.At(4), 3528, 3589)
}

func isNaranJa(ccDigits digits) bool {
	return ccDigits.At(6) == 589562
}

func isMaestro(c *CreditCard, ccDigits digits) bool {
	return matchesValue(ccDigits.At(4), []int{5018, 5020, 5038, 5612, 5893, 6304, 6759, 6761, 6762, 6763, 6390}) ||
		c.Number[:3] == "0604"
}

func isDankort(ccDigits digits) bool {
	return ccDigits.At(4) == 5019
}

func isMasterCard(ccDigits digits) bool {
	return isInBetween(ccDigits.At(2), 51, 55) || isInBetween(ccDigits.At(6), 222100, 272099)
}

func isVisaElectron(ccDigits digits) bool {
	return matchesValue(ccDigits.At(4), []int{4026, 4405, 4508, 4844, 4913, 4917}) || ccDigits.At(6) == 417500
}

func isVisa(ccDigits digits) bool {
	return ccDigits.At(1) == 4
}

func isAura(ccDigits digits) bool {
	return ccDigits.At(2) == 50
}
