package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("timeFile.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var floats []float64

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var f float64
		if err := decoder.Decode(&f); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}
		floats = append(floats, f)
	}

	fmt.Println("Decoded floats:", floats)
}
