package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

const cacheControl = "Cache-Control"
const ccNoCache = "no-cache"
const ccNoStore = "no-store"
const ccPrivate = "private"

var maxAgeReg = regexp.MustCompile(`max-age=(\d+)`)
var sharedMaxAgeReg = regexp.MustCompile(`s-maxage=(\d+)`)

func parseCacheControl(cc string) (cache bool, duration time.Duration) {
	if cc == ccPrivate || cc == ccNoCache || cc == ccNoStore {
		return false, 0
	}

	if cc == "" {
		return true, 1 * time.Hour
	}

	directives := strings.Split(cc, ",")
	for _, directive := range directives {
		directive = strings.ToLower(directive)
		age := maxAgeReg.FindStringSubmatch(directive)
		if len(age) > 0 {
			d, err := strconv.Atoi(age[1])
			if err != nil {
				return false, 0
			}

			cache = true
			duration = time.Duration(d) * time.Second
		}

		age = sharedMaxAgeReg.FindStringSubmatch(directive)
		if len(age) > 0 {
			d, err := strconv.Atoi(age[1])
			if err != nil {
				return false, 0
			}

			cache = true
			duration = time.Duration(d) * time.Second
		}
	}

	return
}
