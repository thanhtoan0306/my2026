package com.my2026.ssrapk

import android.content.Context
import android.os.Build
import fi.iki.elonen.NanoHTTPD
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone

data class PageData(
    val title: String,
    val message: String,
    val name: String,
    val renderedAt: String,
    val runtime: String,
    val platform: String,
)

class SsrServer(
    port: Int,
    private val context: Context,
) : NanoHTTPD(port) {

    override fun serve(session: IHTTPSession): Response {
        val uri = session.uri
        return when {
            uri == "/" || uri.startsWith("/?") -> serveIndex(session)
            uri == "/static/style.css" -> serveAsset("static/style.css", "text/css")
            else -> newFixedLengthResponse(Response.Status.NOT_FOUND, MIME_PLAINTEXT, "Not found")
        }
    }

    private fun serveIndex(session: IHTTPSession): Response {
        val params = session.parms
        val name = params["name"]?.trim()?.ifEmpty { null } ?: "World"
        val data = PageData(
            title = "SSR Apk",
            message = "Hello, $name!",
            name = name,
            renderedAt = isoNow(),
            runtime = "Kotlin ${KotlinVersion.CURRENT} / NanoHTTPD",
            platform = "Android ${Build.VERSION.RELEASE} (API ${Build.VERSION.SDK_INT}) / ${Build.SUPPORTED_ABIS.firstOrNull() ?: Build.CPU_ABI}",
        )
        val html = renderTemplate(loadAsset("templates/index.html"), data)
        return newFixedLengthResponse(Response.Status.OK, "text/html; charset=utf-8", html)
    }

    private fun serveAsset(path: String, mime: String): Response {
        val body = loadAsset(path)
        return newFixedLengthResponse(Response.Status.OK, mime, body)
    }

    private fun loadAsset(path: String): String {
        context.assets.open(path).use { input ->
            return input.bufferedReader().readText()
        }
    }

    private fun renderTemplate(template: String, data: PageData): String {
        val values = mapOf(
            "Title" to escapeHtml(data.title),
            "Message" to escapeHtml(data.message),
            "Name" to escapeHtml(data.name),
            "RenderedAt" to escapeHtml(data.renderedAt),
            "Runtime" to escapeHtml(data.runtime),
            "Platform" to escapeHtml(data.platform),
        )
        var html = template
        for ((key, value) in values) {
            html = html.replace("{{$key}}", value)
        }
        return html
    }

    private fun escapeHtml(text: String): String {
        return text
            .replace("&", "&amp;")
            .replace("<", "&lt;")
            .replace(">", "&gt;")
            .replace("\"", "&quot;")
    }

    private fun isoNow(): String {
        val fmt = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ssXXX", Locale.US)
        fmt.timeZone = TimeZone.getDefault()
        return fmt.format(Date())
    }
}
