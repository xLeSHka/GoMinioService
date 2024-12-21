package utils

import "time"
//выполняет функцию attemt раз с задержкой между ними в delay*i*2 
func DoWithTries(fn func() error, attemt int, delay time.Duration) (err error) {
	for i := 0; i < attemt; i++ {
		if err = fn(); err != nil {
			time.Sleep(delay * time.Duration(i*2))
			i++
			continue
		}
		return nil
	}
	return
}
