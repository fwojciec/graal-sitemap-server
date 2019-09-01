package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	siteURL  = "https://graalagency.com"
	endpoint = "http://localhost:4000/query"
	query    = `{ "query": "{ clients: clients { slug } authors: authors { slug } }" }`
	port     = ":1234"
)

func main() {
	http.HandleFunc("/", handler)
	log.Printf("🚀 Sitemap server started at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func makeStaticPages(includeAuthors bool) []string {
	var staticPages []string
	if includeAuthors {
		staticPages = []string{"/", "/authors", "/clients", "/mailing-list", "/about-us", "/contact"}
	} else {
		staticPages = []string{"/", "/clients", "/mailing-list", "/about-us", "/contact"}
	}
	return staticPages
}

func buildStatic(staticPages []string) []*url {
	urls := make([]*url, 0)
	for _, page := range staticPages {
		locEn := strings.TrimRight(fmt.Sprintf("%s/en%s", siteURL, page), "/")
		locPl := strings.TrimRight(fmt.Sprintf("%s/pl%s", siteURL, page), "/")
		links := []*link{
			&link{Rel: "alternate", HrefLang: "en", Href: locEn},
			&link{Rel: "alternate", HrefLang: "pl", Href: locPl},
		}
		urlEN := &url{Loc: locEn, ChangeFreq: "monthly", Priority: 0.5, Links: links}
		urlPL := &url{Loc: locPl, ChangeFreq: "monthly", Priority: 0.5, Links: links}
		urls = append(urls, urlEN, urlPL)
	}
	return urls
}

func buildDynamic(prefix string, slugs []string) []*url {
	urls := make([]*url, 0)
	for _, slug := range slugs {
		locEn := fmt.Sprintf("%s/en/%s/%s", siteURL, prefix, slug)
		locPl := fmt.Sprintf("%s/pl/%s/%s", siteURL, prefix, slug)
		links := []*link{
			&link{Rel: "alternate", HrefLang: "en", Href: locEn},
			&link{Rel: "alternate", HrefLang: "pl", Href: locPl},
		}
		urlEN := &url{Loc: locEn, ChangeFreq: "weekly", Priority: 0.7, Links: links}
		urlPL := &url{Loc: locPl, ChangeFreq: "weekly", Priority: 0.7, Links: links}
		urls = append(urls, urlEN, urlPL)
	}
	return urls
}

func handler(w http.ResponseWriter, r *http.Request) {
	as, cs, err := getSlugs()
	if err != nil {
		log.Printf("error: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	sp := makeStaticPages(len(as) > 0)
	us := newURLSet(
		buildStatic(sp),
		buildDynamic("authors", as),
		buildDynamic("clients", cs),
	)
	output, err := xml.MarshalIndent(&us, "  ", "    ")
	if err != nil {
		log.Printf("error: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "text/xml")
	w.Write([]byte(xml.Header))
	w.Write(output)
	w.Write([]byte("\n"))
}
