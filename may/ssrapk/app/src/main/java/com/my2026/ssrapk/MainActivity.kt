package com.my2026.ssrapk

import android.annotation.SuppressLint
import android.os.Bundle
import android.webkit.WebResourceRequest
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.appcompat.app.AppCompatActivity
import fi.iki.elonen.NanoHTTPD
import java.io.IOException
import java.net.ServerSocket

class MainActivity : AppCompatActivity() {

    private var server: SsrServer? = null
    private var baseUrl: String? = null

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val port = findFreePort()
        val ssrServer = SsrServer(port, applicationContext)
        try {
            ssrServer.start(NanoHTTPD.SOCKET_READ_TIMEOUT, false)
        } catch (e: IOException) {
            throw RuntimeException("Failed to start SSR server", e)
        }
        server = ssrServer
        baseUrl = "http://127.0.0.1:$port/"

        val webView = WebView(this)
        setContentView(webView)

        webView.settings.javaScriptEnabled = true
        webView.settings.domStorageEnabled = true
        webView.webViewClient = object : WebViewClient() {
            override fun shouldOverrideUrlLoading(
                view: WebView,
                request: WebResourceRequest,
            ): Boolean {
                val url = request.url.toString()
                val base = baseUrl ?: return false
                if (url.startsWith(base) || url.startsWith("http://127.0.0.1:")) {
                    return false
                }
                return true
            }
        }
        webView.loadUrl(baseUrl!!)
    }

    override fun onDestroy() {
        server?.stop()
        server = null
        super.onDestroy()
    }

    private fun findFreePort(): Int {
        ServerSocket(0).use { return it.localPort }
    }
}
