//go:build !solution

package speller

import (
	"log"
	"strings"
)

var (
	numbersUnder20 = [...]string{
		"",
		"one",
		"two",
		"three",
		"four",
		"five",
		"six",
		"seven",
		"eight",
		"nine",
		"ten",
		"eleven",
		"twelve",
		"thirteen",
		"fourteen",
		"fifteen",
		"sixteen",
		"seventeen",
		"eighteen",
		"nineteen",
	}

	tens = [...]string{
		"",
		"",
		"twenty",
		"thirty",
		"forty",
		"fifty",
		"sixty",
		"seventy",
		"eighty",
		"ninety",
	}
)

func spellUnder100(n int) string {
	if n < 20 {
		return numbersUnder20[n]
	}

	tensPart := tens[n/10]
	if n%10 == 0 {
		return tensPart
	}

	return tensPart + "-" + numbersUnder20[n%10]
}

func spellUnder1000(n int) string {
	var (
		hundredsPart string
		remPart      string
	)

	hundreds := n / 100
	if hundreds > 0 {
		hundredsPart = numbersUnder20[hundreds] + " hundred"
	}

	rem := n % 100
	if rem > 0 {
		remPart = spellUnder100(rem)
	}

	switch {
	case hundredsPart != "" && remPart != "":
		return hundredsPart + " " + remPart
	case hundredsPart != "":
		return hundredsPart
	default:
		return remPart
	}
}

func Spell(n int64) string {
	if n == 0 {
		return "zero"
	}

	if n < 0 {
		return "minus " + Spell(-n)
	}

	powersOf10 := [...]struct {
		value int64
		name  string
	}{
		{1_000_000_000, "billion"},
		{1_000_000, "million"},
		{1_000, "thousand"},
		{1, ""},
	}

	var res strings.Builder
	for _, power := range powersOf10 {
		quotient := int(n / power.value)
		if quotient == 0 {
			continue
		}

		if res.Len() > 0 {
			if err := res.WriteByte(' '); err != nil {
				log.Fatal(err)
			}
		}

		var part string
		if power.name != "" {
			part = spellUnder1000(quotient) + " " + power.name
		} else {
			part = spellUnder1000(quotient)
		}

		if _, err := res.WriteString(part); err != nil {
			log.Fatal(err)
		}

		n %= power.value
	}

	return res.String()
}
