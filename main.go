package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	inFileName  = "in.txt"
	outFileName = "out.txt"
)

func executablePath() string {
	executablePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return filepath.Dir(executablePath)
}

func inFile() string {
	return filepath.Join(executablePath(), inFileName)
}

func outFile() string {
	return filepath.Join(executablePath(), outFileName)
}

func stdinToFile(wg *sync.WaitGroup) {
	defer wg.Done()

	stdinScanner := bufio.NewScanner(os.Stdin)
	f, err := os.Create(inFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open %s: %v\n", inFile(), err)
		return
	}
	defer f.Close()
	defer os.Remove(inFile())

	w := bufio.NewWriter(f)
	defer w.Flush()

	for stdinScanner.Scan() {
		line := stdinScanner.Text()
		_, err := w.WriteString(line + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "faild to write to %s: %v\n", inFile(), err)
			break
		}
		w.Flush() // リアルタイム性を保つため、書き込むたびにFlush
	}
	if err := stdinScanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "faild to read stdin: %v\n", err)
	}
}

func fileToStdout(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// 1秒ごとにファイルを読み込む
		func() {
			f, err := os.Open(outFile())
			if err != nil {
				return
			}
			defer f.Close()
			defer os.Remove(outFile())

			inScanner := bufio.NewScanner(f)

			for inScanner.Scan() {
				line := inScanner.Text()
				if line != "" {
					fmt.Println(line)
				}
			}
			if err := inScanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", outFile(), err)
			}
		}()
		time.Sleep(1 * time.Second)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go stdinToFile(&wg)
	go fileToStdout(&wg)

	wg.Wait()
}
