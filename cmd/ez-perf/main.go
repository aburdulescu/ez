package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	args := os.Args[1:]
	if len(args) < 1 {
		return fmt.Errorf("missing csv file")
	}
	csvfilepath := args[0]
	f, err := os.Open(csvfilepath)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}
	t := make([]float64, 0, len(records))
	m := make([]float64, 0, len(records))
	for i := range records {
		elapsedTime, err := strconv.ParseFloat(records[i][0], 64)
		if err != nil {
			log.Println(err)
		}
		t = append(t, elapsedTime)
		mem, err := strconv.ParseFloat(records[i][1], 64)
		if err != nil {
			log.Println(err)
		}
		m = append(m, mem)
	}
	sort.Slice(t, func(i, j int) bool { return t[i] < t[j] })
	sort.Slice(m, func(i, j int) bool { return m[i] < m[j] })

	fmt.Println("Time:")
	printStats(t)
	fmt.Println("")

	fmt.Println("Memory:")
	printStats(m)

	return nil
}

func printStats(data []float64) {
	best := data[0]
	fmt.Println("best\t", best)

	worst := data[len(data)-1]
	fmt.Println("worst\t", worst)

	sum := 0.0
	for i := range data {
		sum += data[i]
	}
	fmt.Println("avg\t", sum/float64(len(data)))

	p50 := int(50.0 / 100.0 * float64(len(data)))
	fmt.Println("50p\t", data[p50])

	p90 := int(90.0 / 100.0 * float64(len(data)))
	fmt.Println("90p\t", data[p90])
}
