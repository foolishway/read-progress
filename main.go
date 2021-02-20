package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	COLOR_NORMAL = "\033[0m"
	COLOR_RED    = "\033[1;33m"
	DELI         = "//readed"
)

func Red(content string) string {
	return fmt.Sprintf("%s%s%s", COLOR_RED, content, COLOR_NORMAL)
}

type Summary struct {
	totalLines int64
	readLines  int64
	progress   float32
}

func (s *Summary) Print() {
	fmt.Fprintf(os.Stdout, "total %s lines;\nreaded %s lines;\nprogress: %s\n",
		Red(strconv.FormatInt(s.totalLines, 10)),
		Red(strconv.FormatInt(s.readLines, 10)),
		Red(fmt.Sprintf("%0.2f", float32(s.progress*100))+"%"),
	)
}

var pipe = make(chan string, runtime.NumCPU())
var wg sync.WaitGroup

func main() {
	s := &Summary{}

	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(s)
	}

	filepath.Walk("./", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			dotIndex := strings.Index(info.Name(), ".")
			if dotIndex == -1 {
				return nil
			}

			ext := info.Name()[dotIndex+1:]
			if ext != "go" && ext != "GO" {
				return nil
			}

			pipe <- path
		}
		return nil
	})

	close(pipe)
	wg.Wait()
	s.progress = float32(s.readLines) / float32(s.totalLines)
	s.Print()
}

//foolishway
func worker(s *Summary) {
	for path := range pipe {
		f, err := os.Open(path)
		if err != nil {
			log.Printf("Read %s error: %v", path, err)
			continue
		}

		var readed bool

		b := bufio.NewScanner(f)
		for b.Scan() {
			line := b.Text()

			if strings.TrimSpace(line) == DELI {
				if readed {
					readed = false
				} else {
					readed = true
				}
				atomic.AddInt64(&s.readLines, 1)
			}

			if readed {
				atomic.AddInt64(&s.readLines, 1)
			}

			atomic.AddInt64(&s.totalLines, 1)
		}
	}
	wg.Done()

}
