package main

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	ErrUnsupportedRecordType = fmt.Errorf("unsupported record type")
)

type strings []string

func (ss *strings) String() string {
	return fmt.Sprint([]string(*ss))
}

func (ss *strings) Set(value string) error {
	if *ss == nil {
		*ss = make(strings, 1)
	} else {
		nss := make(strings, len(*ss)+1)
		copy(nss, *ss)
		*ss = nss
	}
	(*ss)[len(*ss)-1] = value
	return nil
}

type TTL int64

var ttlRegexp = regexp.MustCompile(`(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?`)

func (t *TTL) UnmarshalJSON(data []byte) error {
	var ttl int64
	unquoted := string(data[1 : len(data)-1])
	matches := ttlRegexp.FindStringSubmatch(unquoted)
	if matches[1] != "" {
		weeks, _ := strconv.ParseInt(matches[1], 10, 64)
		ttl += weeks * 7 * 24 * 60 * 60
	}
	if matches[2] != "" {
		days, _ := strconv.ParseInt(matches[2], 10, 64)
		ttl += days * 24 * 60 * 60
	}
	if matches[3] != "" {
		hours, _ := strconv.ParseInt(matches[3], 10, 64)
		ttl += hours * 60 * 60
	}
	if matches[4] != "" {
		minutes, _ := strconv.ParseInt(matches[4], 10, 64)
		ttl += minutes * 60
	}
	if matches[5] != "" {
		seconds, _ := strconv.ParseInt(matches[5], 10, 64)
		ttl += seconds
	}

	*t = TTL(ttl)
	return nil
}

type Record struct {
	ID   string `json:".id"`
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  TTL    `json:"ttl"`

	//targets
	Address      string `json:"address"`
	CName        string `json:"cname"`
	MXPreference string `json:"mx-preference"`
	MXExchange   string `json:"mx-exchange"`
	NS           string `json:"ns"`
	SrvPriority  string `json:"srv-priority"`
	SrvWeight    string `json:"srv-weight"`
	SrvPort      string `json:"srv-port"`
	SrvTarget    string `json:"srv-target"`
	Text         string `json:"text"`
}
