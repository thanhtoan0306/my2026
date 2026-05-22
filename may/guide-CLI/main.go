package main

import (
    "html/template"
    "log"
    "net/http"
    "path/filepath"
    "strings"
)

type PageData struct {
    User    string
    Host    string
    Port    string
    Command string
    Result  string
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", indexHandler)
    mux.HandleFunc("/command", commandHandler)
    mux.HandleFunc("/run", runHandler)
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join("static")))))

    log.Println("Starting guide-CLI on http://localhost:8080")
    err := http.ListenAndServe(":8080", mux)
    if err != nil {
        log.Fatal(err)
    }
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    t, err := template.ParseFiles(filepath.Join("templates", tmpl))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    data := PageData{Port: "22"}
    renderTemplate(w, "index.html", data)
}

func commandHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }
    data := makePageData(r)
    renderTemplate(w, "snippet.html", data)
}

func runHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }
    data := makePageData(r)
    data.Result = runSimulation(data.Command)
    renderTemplate(w, "result.html", data)
}

func makePageData(r *http.Request) PageData {
    user := strings.TrimSpace(r.FormValue("user"))
    host := strings.TrimSpace(r.FormValue("host"))
    port := strings.TrimSpace(r.FormValue("port"))
    if port == "" {
        port = "22"
    }
    if user == "" {
        user = "root"
    }
    command := buildSSHCommand(user, host, port)
    return PageData{User: user, Host: host, Port: port, Command: command}
}

func buildSSHCommand(user, host, port string) string {
    if host == "" {
        return "ssh user@device-ip -p 22"
    }
    if user == "" {
        user = "root"
    }
    if port == "" {
        port = "22"
    }
    return "ssh " + user + "@" + host + " -p " + port
}

func runSimulation(command string) string {
    if strings.TrimSpace(command) == "" {
        return "Enter a device address and click Run to preview the SSH command."
    }
    return "Preview only: this would run the command:\n" + command
}
