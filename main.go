package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"

	"github.com/corona10/goimagehash"
)

type HashResult struct {
	FilePath string
	Hash     *goimagehash.ImageHash
}

func calculateHash(filePaths chan string, hashCh chan<- HashResult) {
	// Range to continously receive from the channel filePaths, looping until it is closed
	for filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("[!] Error opening file %s: %v\n", filePath, err)
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			fmt.Printf("[!] Error decoding image %s: %v\n", filePath, err)
			return
		}

		hash, err := goimagehash.DifferenceHash(img)
		if err != nil {
			fmt.Printf("[!] Error calculating hash for %s: %v\n", filePath, err)
			return
		}
		// Send perceptual hash result to the channel as a struct
		hashCh <- HashResult{FilePath: filePath, Hash: hash}
	}
}

func mvDups(dirPath string, maxWorkers int, quiet bool) error {
	// Buffered channel - can hold the value of maxWorkers before the sender will block
	filePaths := make(chan string, maxWorkers)
	hashCh := make(chan HashResult)
	// Create an empty map where keys are hash values (string) and values are slices of file paths ([]string)
	imgHashes := make(map[string][]string)

	imgPaths, err := getImgPaths(dirPath)
	if err != nil {
		return err
	}

	// Create worker goroutines as a resource pool
	for i := 0; i < cap(filePaths); i++ {
		go calculateHash(filePaths, hashCh)
	}

	// Send to the workers in a separate goroutine
	// The result-gathering loop needs to start before more than maxWorkers items of work can continue
	go func() {
		for _, path := range imgPaths {
			filePaths <- path
		}
		close(filePaths) // Close filePaths channel after sending all paths
	}()

	// Result-gathering loop that receives on the results channel until calculateHash goroutines are done
	for range imgPaths {
		result := <-hashCh
		imgHashes[result.Hash.ToString()] = append(imgHashes[result.Hash.ToString()], result.FilePath)
	}

	close(hashCh)

	// Process each group of hashed images
	dupsDir := filepath.Join(dirPath, "hashed")
	if err := os.Mkdir(dupsDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}

	for hash, paths := range imgHashes {
		if len(paths) > 0 {
			sort.Strings(paths)
			// Use the hash of the first file as the folder name
			hashDir := filepath.Join(dupsDir, hash[2:])
			if err := os.Mkdir(hashDir, os.ModePerm); err != nil && !os.IsExist(err) {
				return err
			}
			// Move similar to the hash-named folder
			for _, p := range paths {
				newPath := filepath.Join(hashDir, filepath.Base(p))
				if !quiet {
					fmt.Printf("[-] Moving: %s to %s\n", p, newPath)
				}
				if err := os.Rename(p, newPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getImgPaths(dirPath string) ([]string, error) {
	var imgPaths []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(info.Name()) == ".jpg" || filepath.Ext(info.Name()) == ".jpeg" || filepath.Ext(info.Name()) == ".png") {
			imgPaths = append(imgPaths, path)
		}
		return nil
	})

	return imgPaths, err
}

func main() {
	dir := flag.String("dir", "", "Images folder path")
	quiet := flag.Bool("quiet", false, "If true, won't print the moved files (default false)")
	maxWorkers := flag.Int("workers", 100, "Number of workers to run concurrently")
	flag.Parse()

	if *dir == "" {
		fmt.Printf("Usage: %s -dir <folder-path>\n", os.Args[0])
		os.Exit(1)
	}

	err := mvDups(*dir, *maxWorkers, *quiet)
	if err != nil {
		fmt.Printf("[!] Error: %v\n", err)
	}
}
