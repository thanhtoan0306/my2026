package com.example.hellowatch

import android.app.Activity
import android.graphics.Color
import android.os.Bundle
import android.view.Gravity
import android.widget.TextView

class MainActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(
            TextView(this).apply {
                text = "Hello World"
                textSize = 24f
                setTextColor(Color.WHITE)
                gravity = Gravity.CENTER
            }
        )
    }
}
