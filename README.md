# ton-dice-web-fetcher
Fetching game results and send them to the dice server to update.

Fetch game results:
 - ton-api grpc 

Update bets:
 - ton-dice-web-server grpc


## build 
```
export GITHUB_TOKEN = 'token'
docker build --build-arg GITHUB_TOKEN="$GITHUB_TOKEN" -t dice-fetcher .
```

## run (develop)
```docker run --name dice-fetcher --network dice-network -e CONTRACT_ADDR=kf_P3yXtu1ab1CBEl6rG2jb_LtCPIElEoqCZTiffBOte0uLL -e STORAGE_HOST=dice-server -e STORAGE_PORT=5300 -e TON_API_HOST=ton-api -e TON_API_PORT=5400 -d dice-fetcher```
 
## ENV VARS
    * CONTRACT_ADDR - Dice contract address in the TON network, required variable, no default value.
    * STORAGE_HOST - Host of the 'ton-dice-web-server' service, requred variable, no default value.
    * STORAGE_PORT - Port of the 'ton-dice-web-server' service, default value is '5300'.
    * TON_API_HOST - Host of the 'ton-api' service, requred variable, no default value.
    * TON_API_PORT - Port of the 'ton-api' service, default value is '5400'.
