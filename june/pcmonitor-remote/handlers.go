package main

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

var monitorToken string

type pageData struct {
	Token       string
	Hostname    string
	Now         time.Time
	Snap        snapshot
	PollEveryMs int
	ProcSort    string
}

func handleIndex(st *store, w http.ResponseWriter, r *http.Request) {
	snap := st.get()
	if snap.Hostname == "" {
		snap.Hostname = "this-pc"
	}
	sortKey := procSortFromReq(r)
	data := pageData{
		Token:       monitorToken,
		Hostname:    snap.Hostname,
		Now:         time.Now(),
		Snap:        snap,
		PollEveryMs: 2000,
		ProcSort:    sortKey,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.ExecuteTemplate(w, "index", data)
}

func handleSummaryFragment(st *store, w http.ResponseWriter, r *http.Request) {
	snap := st.get()
	data := pageData{
		Token:    monitorToken,
		Now:      time.Now(),
		Snap:     snap,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.ExecuteTemplate(w, "summary_fragment", data)
}

func handleProcsFragment(st *store, w http.ResponseWriter, r *http.Request) {
	snap := st.get()

	sortKey := procSortFromReq(r)
	procs := append([]procStat(nil), snap.Procs...)
	switch sortKey {
	case "mem":
		sort.Slice(procs, func(i, j int) bool { return procs[i].RSS > procs[j].RSS })
	default:
		sortKey = "cpu"
		sort.Slice(procs, func(i, j int) bool { return procs[i].CPU > procs[j].CPU })
	}
	snap.Procs = procs

	data := pageData{
		Token:    monitorToken,
		Now:      time.Now(),
		Snap:     snap,
		ProcSort: sortKey,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.ExecuteTemplate(w, "procs_fragment", data)
}

func procSortFromReq(r *http.Request) string {
	v := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
	if v == "mem" {
		return "mem"
	}
	return "cpu"
}
