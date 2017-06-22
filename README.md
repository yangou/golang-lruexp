# lruexp
golang lru cache with randomized expiry on objects


## usage

```golang
package main

import (
	"fmt"
	"github.com/yangou/golang-lruexp"
	"time"
)

func main() {
	// sync cache

	// size of cache: 1000
	// default expiry: 5 seconds, (0 means never expiry by default)
	// randomize by 1000 milliseconds
	// enable cache on nil object (disabled by default)
	syncCache, _ := lruexp.NewSyncCache(1000, 5*time.Second, 1000*time.Millisecond, true)

	// key of cached value: myKey
	// expiry other than default: 1 second
	// fallback function returns string
	value, _ := syncCache.FetchWithFunc("myKey", 1*time.Second, func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return "cached value", nil
	})

	fmt.Println(value)

	//////////////////////////////////////

	// async cache

	// size of cache: 1000
	// default expiry: 5 seconds, (0 means never expiry by default)
	// randomize by 1000 milliseconds
	// enable cache on nil object (disabled by default)
	// error handler: prints "error"
	asyncCache, _ := lruexp.NewAsyncCache(1000, 5*time.Second, 1000*time.Millisecond, true, func(error) {
		fmt.Println("error")
	})

	// key of cached value: myKey
	// expiry other than default: 1 second
	// fallback function returns string
	// error handler: prints "ignore error"
	value, _ = asyncCache.FetchWithFunc("myKey", 1*time.Second, func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return "cached value", fmt.Errorf("random error")
	}, func(error) {
		fmt.Println("ignore error")
	})

	fmt.Println(value)
}

```

## expiry table

| default expiry  | individual expiry  | final expriy  |
|---|---|---|
|= 0   |> 0   | individual expiry  |
|= 0   |= 0   | no expiry  |
|= 0   |< 0   | no expiry  |
|> 0   |> 0   | individual expiry  |
|> 0   |= 0   | default expiry  |
|> 0   |< 0   | no expiry |
