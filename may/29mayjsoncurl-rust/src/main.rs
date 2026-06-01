use std::path::PathBuf;
use std::process::Stdio;
use std::time::Duration;

use axum::extract::{Form, State};
use axum::response::Html;
use axum::routing::get;
use axum::Router;
use minijinja::context;
use minijinja::{AutoEscape, Environment};
use regex::Regex;
use serde::{Deserialize, Serialize};
use tokio::process::Command;
use tokio::time::timeout;

const HOST: &str = "127.0.0.1";
const PORT: u16 = 3030;
const CURL_TIMEOUT: Duration = Duration::from_secs(30);

#[derive(Clone, Serialize)]
struct RunResult {
    ok: bool,
    status: String,
    body: String,
    error: Option<String>,
}

#[derive(Deserialize)]
struct CurlForm {
    curl: String,
}

fn templates_dir() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("templates")
}

fn validate_curl(cmd: &str) -> Option<String> {
    let text = cmd.trim();
    if text.is_empty() {
        return Some("Paste a curl command first.".into());
    }

    static CURL_START: std::sync::OnceLock<Regex> = std::sync::OnceLock::new();
    static UNSAFE: std::sync::OnceLock<Regex> = std::sync::OnceLock::new();

    let curl_start = CURL_START.get_or_init(|| Regex::new(r"(?i)^\s*curl\b").unwrap());
    let unsafe_pat = UNSAFE.get_or_init(|| {
        Regex::new(r"(?i)(?:^|[\s;])(?:rm\s|sudo\s|curl\s+.*\|\s*sh|>\s*/|/dev/tcp)").unwrap()
    });

    if !curl_start.is_match(text) {
        return Some("Command must start with curl.".into());
    }
    if unsafe_pat.is_match(text) {
        return Some("Command blocked for safety.".into());
    }
    if has_shell_chaining(text) {
        return Some("Shell chaining (&&, ||, ;, $(), etc.) is not allowed.".into());
    }
    None
}

fn has_shell_chaining(text: &str) -> bool {
    if text.contains("&&")
        || text.contains("||")
        || text.contains('`')
        || text.contains("$(")
        || text.contains("${")
        || text.contains("<(")
        || text.contains(">(")
    {
        return true;
    }
    text.rfind(';').is_some_and(|idx| !text[idx + 1..].trim().is_empty())
}

fn sort_json_value(value: serde_json::Value) -> serde_json::Value {
    match value {
        serde_json::Value::Object(map) => {
            let mut keys: Vec<_> = map.keys().cloned().collect();
            keys.sort();
            let mut sorted = serde_json::Map::new();
            for key in keys {
                if let Some(v) = map.get(&key) {
                    sorted.insert(key, sort_json_value(v.clone()));
                }
            }
            serde_json::Value::Object(sorted)
        }
        serde_json::Value::Array(items) => serde_json::Value::Array(
            items
                .into_iter()
                .map(sort_json_value)
                .collect::<Vec<_>>(),
        ),
        other => other,
    }
}

fn beautify_body(raw: &str) -> String {
    let stripped = raw.trim();
    if stripped.is_empty() {
        return String::new();
    }

    match serde_json::from_str::<serde_json::Value>(stripped) {
        Ok(value) => {
            let sorted = sort_json_value(value);
            match serde_json::to_string_pretty(&sorted) {
                Ok(pretty) => format!("{pretty}\n"),
                Err(_) => raw.to_string(),
            }
        }
        Err(_) => raw.to_string(),
    }
}

async fn run_curl(cmd: &str) -> RunResult {
    let child = Command::new("/bin/bash")
        .arg("-lc")
        .arg(cmd)
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .kill_on_drop(true)
        .spawn();

    let child = match child {
        Ok(c) => c,
        Err(e) => {
            return RunResult {
                ok: false,
                status: "error".into(),
                body: String::new(),
                error: Some(e.to_string()),
            };
        }
    };

    let output = timeout(CURL_TIMEOUT, child.wait_with_output()).await;

    match output {
        Ok(Ok(out)) => {
            let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
            let stderr = String::from_utf8_lossy(&out.stderr).into_owned();
            let combined = if stdout.trim().is_empty() {
                stderr.clone()
            } else {
                stdout
            };
            let pretty = beautify_body(&combined);
            let code = out.status.code().unwrap_or(-1);

            if !out.status.success() && pretty.trim().is_empty() {
                return RunResult {
                    ok: false,
                    status: format!("exit {code}"),
                    body: pretty,
                    error: Some(
                        if stderr.trim().is_empty() {
                            format!("curl exited with code {code}")
                        } else {
                            stderr.trim().to_string()
                        },
                    ),
                };
            }

            RunResult {
                ok: out.status.success(),
                status: format!("exit {code}"),
                body: pretty,
                error: if !out.status.success() && !stderr.trim().is_empty() {
                    Some(stderr.trim().to_string())
                } else {
                    None
                },
            }
        }
        Ok(Err(e)) => RunResult {
            ok: false,
            status: "error".into(),
            body: String::new(),
            error: Some(e.to_string()),
        },
        Err(_) => RunResult {
            ok: false,
            status: "timeout".into(),
            body: String::new(),
            error: Some(format!(
                "Timed out after {}s",
                CURL_TIMEOUT.as_secs()
            )),
        },
    }
}

fn render(
    env: &Environment<'_>,
    curl_text: &str,
    validation_error: Option<&str>,
    result: Option<RunResult>,
) -> Html<String> {
    let tmpl = env.get_template("index.html").expect("template");
    let html = tmpl
        .render(context! {
            curl_text => curl_text,
            validation_error => validation_error,
            result => result,
        })
        .expect("render");
    Html(html)
}

async fn index(State(env): State<Environment<'static>>) -> Html<String> {
    render(&env, "", None, None)
}

async fn run_form(
    State(env): State<Environment<'static>>,
    Form(form): Form<CurlForm>,
) -> Html<String> {
    let curl_text = form.curl;
    if let Some(err) = validate_curl(&curl_text) {
        return render(&env, &curl_text, Some(&err), None);
    }

    let result = run_curl(curl_text.trim()).await;
    render(&env, &curl_text, None, Some(result))
}

#[tokio::main]
async fn main() {
    let mut env = Environment::new();
    env.set_loader(minijinja::path_loader(templates_dir()));
    env.set_auto_escape_callback(|name| {
        if name.ends_with(".html") {
            AutoEscape::Html
        } else {
            AutoEscape::None
        }
    });

    let app = Router::new()
        .route("/", get(index).post(run_form))
        .with_state(env);

    let addr = format!("{HOST}:{PORT}");
    let listener = tokio::net::TcpListener::bind(&addr)
        .await
        .expect("bind");
    println!("Listening on http://{addr}");
    axum::serve(listener, app).await.expect("serve");
}
