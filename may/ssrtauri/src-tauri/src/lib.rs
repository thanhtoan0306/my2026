use axum::{
    extract::Query,
    response::{Html, IntoResponse},
    routing::get,
    Router,
};
use chrono::Local;
use std::collections::HashMap;
use std::sync::mpsc;
use std::time::Duration;
use tauri::{WebviewUrl, WebviewWindowBuilder};

const INDEX_TEMPLATE: &str = include_str!("../templates/index.html");
const STYLE_CSS: &str = include_str!("../static/style.css");

struct PageData {
    title: String,
    message: String,
    name: String,
    rendered_at: String,
    runtime: String,
    platform: String,
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .setup(|app| {
            let (tx, rx) = mpsc::channel();
            std::thread::spawn(move || {
                let rt = tokio::runtime::Runtime::new().expect("tokio runtime");
                rt.block_on(async {
                    let listener = tokio::net::TcpListener::bind("127.0.0.1:0")
                        .await
                        .expect("bind localhost");
                    let port = listener.local_addr().expect("local addr").port();
                    tx.send(port).expect("send port");

                    let app = Router::new()
                        .route("/", get(index_handler))
                        .route("/static/style.css", get(style_handler));

                    axum::serve(listener, app).await.expect("serve");
                });
            });

            let port = rx.recv().expect("receive port");
            std::thread::sleep(Duration::from_millis(150));

            let url = format!("http://127.0.0.1:{}/", port);
            WebviewWindowBuilder::new(
                app,
                "main",
                WebviewUrl::External(
                    url.parse()
                        .expect("valid URL"),
                ),
            )
            .title("SSR Tauri — Hello World")
            .inner_size(720.0, 560.0)
            .resizable(true)
            .build()?;

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

async fn index_handler(Query(params): Query<HashMap<String, String>>) -> Html<String> {
    let name = params
        .get("name")
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
        .unwrap_or("World");

    let data = PageData {
        title: "SSR Tauri".into(),
        message: format!("Hello, {}!", name),
        name: name.to_string(),
        rendered_at: Local::now().to_rfc3339(),
        runtime: format!("Rust {} / Axum / Tauri 2", rustc_version()),
        platform: format!("{} / {}", std::env::consts::OS, std::env::consts::ARCH),
    };

    Html(render_template(data))
}

async fn style_handler() -> impl IntoResponse {
    ([("Content-Type", "text/css; charset=utf-8")], STYLE_CSS)
}

fn render_template(data: PageData) -> String {
    let mut html = INDEX_TEMPLATE.to_string();
    let pairs = [
        ("Title", escape_html(&data.title)),
        ("Message", escape_html(&data.message)),
        ("Name", escape_html(&data.name)),
        ("RenderedAt", escape_html(&data.rendered_at)),
        ("Runtime", escape_html(&data.runtime)),
        ("Platform", escape_html(&data.platform)),
    ];
    for (key, value) in pairs {
        html = html.replace(&format!("{{{{{}}}}}", key), &value);
    }
    html
}

fn escape_html(text: &str) -> String {
    text.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
}

fn rustc_version() -> String {
    option_env!("RUSTC_VERSION")
        .unwrap_or("unknown")
        .to_string()
}
