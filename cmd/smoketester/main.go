package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/alecthomas/kong"
	"github.com/martinlindhe/feng"
	"github.com/martinlindhe/feng/mapper"
	"github.com/martinlindhe/feng/smoketest"
)

var args struct {
	Filename string `kong:"arg" name:"filename" type:"existingfile" help:"Input yaml with file listing."`
}

type measuredExecution struct {
	duration time.Duration
	filename string
}

func main() {

	_ = kong.Parse(&args,
		kong.Name("feng"),
		kong.Description("A binary template reader and data presenter."))

	data, err := ioutil.ReadFile(args.Filename)
	if err != nil {
		log.Fatal(err)
	}

	referenceRoot := "./smoketest/reference"

	err = os.RemoveAll(referenceRoot)
	if err != nil {
		log.Fatal(err)
	}

	smoketests, err := smoketest.UnmarshalData(data)
	if err != nil {
		log.Fatal(err)
	}
	filenames := smoketests.GenerateFilenames(filepath.Dir(args.Filename))

	measures := []measuredExecution{}

	for _, entry := range filenames {
		feng.Green("Start entry %s\n", entry.In)

		started := time.Now()

		fl, err := mapper.MapFileToTemplate(entry.In)
		if err != nil {
			// template don't match, try another
			if _, ok := err.(mapper.EvaluateError); ok {
				log.Println(" failed to evaluate:", err)
			}

			log.Println("MapReader returned err:", err)

			continue
		}

		if len(fl.Structs) == 0 {
			fmt.Println("MapReader failure, skipping")
			continue
		}

		feng.Green("Parsed %s as %s\n\n", entry.In, fl.BaseName)

		data := fl.Present(&mapper.PresentFileLayoutConfig{
			ShowRaw:           true,
			ReportOverlapping: true,
			InUTC:             true})

		timeSpent := time.Since(started)
		measures = append(measures, measuredExecution{
			duration: timeSpent,
			filename: entry.Out,
		})

		filename, _ := filepath.Abs(filepath.Join(referenceRoot, entry.Out))
		path := filepath.Dir(filename)

		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		feng.Green("WRITE ROOT %s, out %s, full %s\n", referenceRoot, entry.Out, filename)
		err = ioutil.WriteFile(filename, []byte(data), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	sort.Slice(measures, func(i, j int) bool {
		return measures[i].duration > measures[j].duration
	})

	// XXX sort by exec speed and show the slowest first
	topMeasures := 25
	if len(measures) > topMeasures {
		measures = measures[:topMeasures]
	}

	fmt.Println("--- the", topMeasures, "slowest mappings were ---")
	for _, measure := range measures {
		fmt.Printf("%v:\t%s\n", measure.duration, measure.filename)
	}

}
