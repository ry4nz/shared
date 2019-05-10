package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"net/http"
)

type Result struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	SIG     string `json:"sig"`
	Comment string `json:"comment"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	csvFile, _ := os.Open("kub.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.LazyQuotes = true
	var results []Result
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		if line[4] == "Failed" {
			results = append(results,
				Result{
					ID:      fmt.Sprintf(`<a href="https://docker.testrail.net/index.php?/tests/view/%s">%s</a>`, line[0][1:], line[0]),
					Title:   line[1],
					SIG:     line[3],
					Comment: line[2],
				},
			)
		}
	}
	//resultsJson, _ := json.Marshal(results)

	page := "<h1>Failed Tests</h1>"
	pageHead := "<table style='width:100%'><tr><th>ID</th><th>Title</th><th>SIG</th>"
	pageTail := "</table>"

	filteredResults := []Result{}

	for _, reason := range []string{
		"no such host",
		"did not become Bound: PersistentVolumeClaim",
		"found but phase is Pending instead of Bound",
	} {
		filteredResults = []Result{}
		page += fmt.Sprintf(`<h2> %s </h2>`, reason)
		page += pageHead

		for _, r := range results {
			if strings.Contains(r.Comment, reason) {
				page += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", r.ID, r.Title, r.SIG)
			} else {
				filteredResults = append(filteredResults, r)
			}
		}
		page += pageTail
		results = filteredResults
	}



	page += "<h2> Uncategorized </h2>"
	page += pageHead
	filteredResults = []Result{}
	for _, r := range results {

			page += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", r.ID, r.Title, r.SIG)

	}
	page += pageTail
	results = filteredResults

	fmt.Fprintf(w, "%s", page)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
