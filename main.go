package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/xpzouying/gofake/errgroup"
)

func main() {
	existsFiles, err := checkTwoFilesSame(".")
	if err != nil {
		log.Fatal(err)
	}

	// logging the result
	processAllFiles(existsFiles)
}

type fileList []string

type result struct {
	path string
	sum  string
}

func checkTwoFilesSame(root string) (map[string]fileList, error) {
	g := errgroup.New()
	paths := make(chan string)

	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			paths <- path
			return nil
		})
	})

	c := make(chan result)
	const numDigesters = 20
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for path := range paths {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				md5str := fmt.Sprintf("%x", md5.Sum(data))
				c <- result{path, md5str}
			}
			return nil
		})
	}
	go func() {
		g.Wait()
		close(c)
	}()

	// md5 --> files
	existsFiles := make(map[string]fileList, 1024)
	for r := range c {
		existsFiles[r.sum] = append(existsFiles[r.sum], r.path)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return existsFiles, nil
}

func processAllFiles(existsFiles map[string]fileList) {
	// logging the result
	var dup bool
	for md5str, files := range existsFiles {
		dup = false
		if len(files) > 1 {
			dup = true
		}

		for _, f := range files {
			log.Printf("%s\t%v\t%s", md5str, dup, f)
		}
	}
}
