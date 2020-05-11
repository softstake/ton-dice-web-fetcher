package fetcher

import (
	"context"
	"fmt"
	"github.com/tonradar/ton-dice-web-fetcher/config"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"strconv"

	api "github.com/tonradar/ton-api/proto"
	store "github.com/tonradar/ton-dice-web-server/proto"
)

const SavedTrxLtFileName = "trxlt.save"

const (
	// bet lifecycle states
	UNSAVED = iota - 1
	SAVED
	SENT
	RESOLVED
)

type Fetcher struct {
	conf          *config.TonWebFetcherConfig
	apiClient     api.TonApiClient
	storageClient store.BetsClient
}

func NewFetcher(conf *config.TonWebFetcherConfig) *Fetcher {
	log.Println("Fetcher init...")
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		withClientUnaryInterceptor(),
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", conf.StorageHost, conf.StoragePort), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	storageClient := store.NewBetsClient(conn)

	conn, err = grpc.Dial(fmt.Sprintf("%s:%d", conf.TonAPIHost, conf.TonAPIPort), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	apiClient := api.NewTonApiClient(conn)

	return &Fetcher{
		conf:          conf,
		apiClient:     apiClient,
		storageClient: storageClient,
	}
}

func (f *Fetcher) FetchResults(lt int64, hash string, depth int) (int64, string) {
	ctx := context.Background()

	fetchTransactionsRequest := &api.FetchTransactionsRequest{
		Address: f.conf.ContractAddr,
		Lt:      lt,
		Hash:    hash,
	}

	fetchTransactionsResponse, err := f.apiClient.FetchTransactions(ctx, fetchTransactionsRequest)
	if err != nil {
		log.Println(err)
		return lt, hash
	}

	transactions := fetchTransactionsResponse.Items
	var trx *api.Transaction

	log.Printf("Fetched %d transactions", len(transactions))

	for _, trx = range transactions {
		log.Printf("Processing a transaction with lt %d and hash %s", trx.TransactionId.Lt, trx.TransactionId.Hash)
		for _, outMsg := range trx.OutMsgs {
			gameResult, err := parseOutMessage(outMsg.Message)
			if err != nil {
				log.Printf("Parse output message failed: %v\n", err)
				continue
			}
			log.Printf("Game with id %d and random roll %d is defined", gameResult.Id, gameResult.RandomRoll)

			isBetResolved, err := isBetResolved(f.storageClient, int32(gameResult.Id))
			if err != nil {
				log.Println(err)
				continue
			}

			if isBetResolved.IsResolved {
				log.Println("The bet is already resolved, proceed to the next transaction...")
				continue
			}

			playerPayout := outMsg.Value
			resolveTrxHash := trx.TransactionId.Hash
			resolveTrxLt := trx.TransactionId.Lt

			req := &store.UpdateBetRequest{
				Id:             int32(gameResult.Id),
				RandomRoll:     int32(gameResult.RandomRoll),
				PlayerPayout:   playerPayout,
				State:          RESOLVED,
				ResolveTrxHash: resolveTrxHash,
				ResolveTrxLt:   resolveTrxLt,
			}

			_, err = f.storageClient.UpdateBet(ctx, req)
			if err != nil {
				log.Printf("Update bet in DB failed: %v\n", err)
				continue
			}
			log.Printf("Bet with id %d successfully updated", gameResult.Id)
		}
	}

	_lt := lt
	_hash := hash
	if len(transactions) > 0 {
		_lt = trx.TransactionId.Lt
		_hash = trx.TransactionId.Hash
		if depth > 0 {
			depth -= 1
			return f.FetchResults(_lt, _hash, depth)
		}
	}

	return _lt, _hash
}

func (f *Fetcher) Start() {
	log.Println("Start fetching game results...")
	for {
		getAccountStateRequest := &api.GetAccountStateRequest{
			AccountAddress: f.conf.ContractAddr,
		}
		getAccountStateResponse, err := f.apiClient.GetAccountState(context.Background(), getAccountStateRequest)
		if err != nil {
			log.Printf("failed get account state: %v\n", err)
			continue
		}

		lt := getAccountStateResponse.LastTransactionId.Lt
		hash := getAccountStateResponse.LastTransactionId.Hash

		log.Printf("current hash: %s, current lt: %d", hash, lt)

		savedTrxLt, err := GetSavedTrxLt(SavedTrxLtFileName)
		if err != nil {
			log.Printf("failed read saved trx lt: %v\n", err)
			return
		}

		if lt > int64(savedTrxLt) {
			err = ioutil.WriteFile(SavedTrxLtFileName, []byte(strconv.Itoa(int(lt))), 0644)
			if err != nil {
				log.Printf("failed write trx lt: %v\n", err)
				return
			}

			f.FetchResults(lt, hash, 3)
		}
	}
}
