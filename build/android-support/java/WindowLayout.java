// build/android-support/java/WindowLayout.java
package com.wails.app;

import android.view.View;
import android.view.Window;
import androidx.core.graphics.Insets;
import androidx.core.view.ViewCompat;
import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsCompat;

final class WindowLayout {
    private WindowLayout() { }

    static void install(Window window, View content) {
        WindowCompat.enableEdgeToEdge(window);
        content.setBackgroundResource(R.color.wails_background);
        int types = WindowInsetsCompat.Type.systemBars()
                | WindowInsetsCompat.Type.displayCutout()
                | WindowInsetsCompat.Type.ime();
        ViewCompat.setOnApplyWindowInsetsListener(content, (view, insets) -> {
            Insets safe = insets.getInsets(types);
            view.setPadding(safe.left, safe.top, safe.right, safe.bottom);
            // The native container owns padding; WebView must still receive inset updates.
            return new WindowInsetsCompat.Builder(insets).setInsets(types, Insets.NONE).build();
        });
        ViewCompat.requestApplyInsets(content);
    }
}
