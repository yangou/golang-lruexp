# lruexp
golang lru cache with randomized expiry on objects


## usage

```golang
package main

import "fmt"
import "time"
import "github.com/yangou/golang-lruexp"

func main() {
	// named cache: myCache
	// size of cache: 1000
	// default expiry: 5 seconds, (0 means never expiry by default)
	// randomize by 1000 milliseconds
	// enable cache on nil object (disabled by default)
	lruexp.New("myCache", 1000, 5*time.Second, 1000, true)

	// key of cached value: myKey
	// expiry if not default: 1 second
	// fallback function is value is missing or expired
	value, _ := lruexp.FetchWithFunc("myCache", "myKey", 1*time.Second, func() (interface{}, error) {
		time.Sleep(3 * time.Second)
		return "cached value", nil
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
