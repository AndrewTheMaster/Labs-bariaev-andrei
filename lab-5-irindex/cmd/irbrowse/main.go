package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"siaod-hw5-irindex/internal/ir"
)

//go:embed templates/*
var templateFS embed.FS

type pageData struct {
	Query      string
	Hits       []ir.HitLine
	HitsN      int
	Elapsed    string
	Err        string
	Doc        *ir.WikiDoc
	DocView        bool
	Debug          ir.DocDebugResult
	HasDebug       bool
	TokenLine      string
	Excerpt    string
	IndexInfo  string
}

func main() {
	indexPath := flag.String("index", "data/index.irx", "mmap index")
	wikiXML := flag.String("wiki-xml", ir.ResolveCorpusPath(), "wiki XML for doc text")
	addr := flag.String("addr", "127.0.0.1:8088", "listen address")
	flag.Parse()

	mi, err := ir.OpenMMapIndex(*indexPath)
	if err != nil {
		log.Fatalf("open index: %v", err)
	}
	defer mi.Close()

	funcs := template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"urlquery": url.QueryEscape,
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}
	tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		docStr := strings.TrimSpace(r.URL.Query().Get("doc"))
		data := pageData{
			Query:     q,
			IndexInfo: fmt.Sprintf("%s — %d docs, %d terms", *indexPath, mi.NumDocs(), mi.Terms()),
		}

		if docStr != "" {
			id, err := strconv.ParseUint(docStr, 10, 32)
			if err != nil {
				data.Err = err.Error()
			} else if *wikiXML == "" {
				data.Err = "укажите -wiki-xml для просмотра текста"
			} else {
				doc, err := ir.FetchWikiDoc(*wikiXML, uint32(id))
				if err != nil {
					data.Err = err.Error()
				} else {
					data.Doc = &doc
					data.DocView = true
					data.Excerpt = ir.TextExcerpt(doc.Text, 400)
					data.TokenLine = ir.FormatTokenLine(doc.Tokens, 40)
					if title := mi.DocTitle(uint32(id)); title != "" {
						data.Doc.Title = title
					}
					if q != "" {
						dbg, err := ir.DocQueryDebug(q, doc.Tokens)
						if err != nil {
							data.Err = err.Error()
						} else {
							data.Debug = dbg
							data.HasDebug = true
						}
					}
				}
			}
		}

		if q != "" && docStr == "" {
			t0 := time.Now()
			_, lines, err := ir.SearchBoolMMapDetailed(mi, q)
			data.Elapsed = time.Since(t0).Round(time.Microsecond).String()
			if err != nil {
				data.Err = err.Error()
			} else {
				data.Hits = lines
				data.HitsN = len(lines)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	log.Printf("irbrowse: http://%s  index=%s  wiki=%s", *addr, *indexPath, *wikiXML)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
