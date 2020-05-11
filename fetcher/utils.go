package fetcher

import (
	"context"
	"encoding/base64"
	"fmt"
	store "github.com/tonradar/ton-dice-web-server/proto"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
)

func parseOutMessage(m string) (*GameResult, error) {
	log.Println("Start parsing an outgoing message...")

	msg, err := base64.StdEncoding.DecodeString(m)
	if err != nil {
		log.Printf("Мessage decode failed: %v\n", err)
		return nil, err
	}

	log.Printf("Decoded message - '%s'", string(msg))

	if len(msg) > 0 {
		r, _ := regexp.Compile(`TONBET.IO - \[#(\d+)] Your number is (\d+), all numbers greater than (\d+) have won.`)
		matches := r.FindStringSubmatch(string(msg))

		if len(matches) > 0 {
			randomRoll, _ := strconv.Atoi(matches[3])
			betID, _ := strconv.Atoi(matches[1])

			return &GameResult{Id: betID, RandomRoll: randomRoll}, nil
		}
		log.Println("Message does not match expected pattern")
	} else {
		log.Println("Message is empty")
	}

	return nil, fmt.Errorf("message is not valid")
}

func isBetResolved(s store.BetsClient, id int32) (*store.IsBetResolvedResponse, error) {
	isBetResolvedReq := &store.IsBetResolvedRequest{
		Id: id,
	}

	resp, err := s.IsBetResolved(context.Background(), isBetResolvedReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetSavedTrxLt(fn string) (int, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return 0, err
	}

	savedTrxLt, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}

	return savedTrxLt, nil
}
