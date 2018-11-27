package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/machinebox/progress"
)

const dlURL = "https://github.com/evansiroky/timezone-boundary-builder/releases/download/%s/timezones.geojson.zip"
const varName = "tzShapeFile"
const template = `// generated by tzshapefilegen; DO NOT EDIT
package gotz

var %s = []byte("%s")
`

func main() {
	_, err := exec.LookPath("mapshaper")
	if err != nil {
		log.Fatalln("Error: mapshaper executable not found in $PATH")
	}

	release := flag.String("release", "2018g", "timezone boundary builder release version")
	flag.Parse()

	resp, err := http.Get(fmt.Sprintf(dlURL, *release))
	if err != nil {
		log.Fatalf("Error: could not download tz shapefile: %v\n", err)
	}
	defer resp.Body.Close()

	r := progress.NewReader(resp.Body)
	size := resp.ContentLength
	wg := sync.WaitGroup{}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		progressChan := progress.NewTicker(ctx, r, size, time.Second)
		fmt.Println("Downloading timezone shape file", *release)
		for p := range progressChan {
			fmt.Printf("\r%v  Remaining...", p.Remaining().Round(time.Second))
		}
		fmt.Println("")
	}()

	releaseZipBuf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(releaseZipBuf, r)
	if err != nil {
		cancel()
		wg.Wait()
		log.Printf("Download failed: %v\n", err)
		return
	}
	wg.Wait()
	cancel()

	releaseZipReadBuf := bytes.NewReader(releaseZipBuf.Bytes())
	z, err := zip.NewReader(releaseZipReadBuf, size)
	if err != nil {
		log.Printf("Could not access zipfile: %v\n", err)
		return
	}
	if len(z.File) == 0 {
		log.Println("Error: release zip file have no files!")
		return
	} else if z.File[0].Name != "dist/combined.json" {
		log.Println("Error: first file in zip file is not dist/combined.json")
		return
	}

	combinedJSONZip, err := z.File[0].Open()
	if err != nil {
		log.Printf("Error: could not read from zip file: %v\n", err)
		return
	}

	currDir, err := os.Getwd()
	if err != nil {
		log.Printf("Error: could not get current dir: %v\n", err)
		return
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Printf("Error: could not create tmp dir: %v\n", err)
		return
	}

	err = os.Chdir(tmpDir)
	if err != nil {
		log.Printf("Error: could not switch to tmp dir: %v\n", err)
		return
	}

	combinedJSON, err := os.Create("./combined.json")
	if err != nil {
		log.Printf("Error: could not create combinedJSON file: %v\n", err)
		return
	}

	_, err = io.Copy(combinedJSON, combinedJSONZip)
	if err != nil {
		combinedJSON.Close()
		log.Printf("Error: could not copy from zip to combined.json: %v\n", err)
		return
	}
	combinedJSON.Close()

	fmt.Println("*** RUNNING MAPSHAPER ***")
	mapshaper := exec.Command("mapshaper", "-i", "combined.json", "-simplify", "visvalingam", "20%", "-o", "reduced.json")
	mapshaper.Stdout = os.Stdout
	mapshaper.Stderr = os.Stderr
	err = mapshaper.Run()
	if err != nil {
		log.Printf("Error: could not run mapshaper: %v\n", err)
		return
	}
	fmt.Println("*** MAPSHAPER FINISHED ***")

	fmt.Println("*** GENERATING GO CODE ***")
	f, err := os.Open("reduced.json")
	if err != nil {
		log.Printf("Error: could not open file: %v\n", err)
		return
	}
	defer f.Close()

	buf := bytes.NewBuffer([]byte{})
	g, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		log.Printf("Error: could not create gzip writer: %v\n", err)
		return
	}
	defer g.Close()

	_, err = io.Copy(g, f)
	if err != nil {
		log.Printf("Error: could not copy data: %v\n", err)
		return
	}
	if err := g.Flush(); err != nil {
		log.Printf("Error: could not flush gzip: %v\n", err)
		return
	}

	data := buf.Bytes()
	hexStr := bytes.NewBuffer([]byte{})
	for i := range data {
		if int(data[i]) < 16 {
			hexStr.WriteString("\\x" + fmt.Sprintf("0%X", data[i]))
		} else {
			hexStr.WriteString("\\x" + fmt.Sprintf("%X", data[i]))
		}
	}
	content := fmt.Sprintf(template, varName, hexStr)

	err = os.Chdir(currDir)
	if err != nil {
		log.Printf("Error: could not switch to previous dir: %v", err)
		return
	}

	fout, err := os.Create("tzshapefile.go")
	if err != nil {
		log.Printf("Error: could not create tzshapefile.go: %v", err)
		return
	}
	defer fout.Close()

	_, err = fout.WriteString(content)
	if err != nil {
		log.Printf("Error: could not write content: %v", err)
		return
	}

	os.RemoveAll(tmpDir)
	fmt.Println("*** ALL DONE, YAY ***")
}
