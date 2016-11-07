package clinic

import (
	"fmt"
	"sync"
	"time"
)

var (
	frames = `|/-\`
)

func (a *App) Spin(msg string, fn func() error) error {
	pattern := "\r%s[%s]%s %s"
	if !stdoutColor {
		err := fn()
		if err != nil {
			fmt.Printf(pattern+"\n", "", "FAIL", "", msg)
			return err
		}

		fmt.Printf(pattern+"\n", "", "OK", "", msg)
		return nil
	}

	m := sync.Mutex{}
	ticker := time.NewTicker(time.Millisecond * 100)
	go func() {
		next := 0
		for _ = range ticker.C {
			m.Lock()
			fmt.Printf(pattern, blue, string(frames[next]), reset, msg)

			next += 1
			if next >= len(frames) {
				next = 0
			}
			m.Unlock()
		}
	}()

	err := fn()
	ticker.Stop()
	m.Lock()
	defer m.Unlock()
	if err != nil {
		fmt.Printf(pattern+"\n", red, "FAIL", reset, msg)
		return err
	}

	fmt.Printf(pattern+"\n", blue, "OK", reset, msg)
	return nil
}
