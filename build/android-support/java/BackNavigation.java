// build/android-support/java/BackNavigation.java
package com.wails.app;

import android.os.Handler;
import android.os.Looper;
import android.webkit.JavascriptInterface;
import android.webkit.WebView;
import androidx.activity.OnBackPressedCallback;
import androidx.appcompat.app.AppCompatActivity;

public final class BackNavigation implements AutoCloseable {
    private final AppCompatActivity activity;
    private final WebView view;
    private final Handler ui = new Handler(Looper.getMainLooper());
    private final OnBackPressedCallback callback;
    private final Runnable timeout = () -> {
        waiting = false;
        android.util.Log.w("SmeggTunerNavigation", "Frontend Back action timed out");
    };
    private int sequence;
    private boolean waiting;
    private boolean closed;

    BackNavigation(AppCompatActivity activity, WebView view) {
        this.activity = activity;
        this.view = view;
        callback = new OnBackPressedCallback(false) {
            @Override public void handleOnBackPressed() { request(); }
        };
        activity.getOnBackPressedDispatcher().addCallback(activity, callback);
        view.addJavascriptInterface(this, "smeggTunerBack");
    }

    @JavascriptInterface public void setEnabled(boolean enabled) {
        ui.post(() -> { if (!closed) callback.setEnabled(enabled); });
    }

    private void request() {
        if (closed || waiting) return;
        waiting = true;
        int token = ++sequence;
        ui.postDelayed(timeout, 15000);
        // Failures consume Back so an unsaved editor cannot accidentally close the app.
        view.evaluateJavascript("Promise.resolve().then(() => window.__smeggTunerBack())"
                + ".then(consumed => smeggTunerBack.complete(" + token + ", consumed !== false),"
                + "() => smeggTunerBack.complete(" + token + ", true));", null);
    }

    @JavascriptInterface public void complete(int token, boolean consumed) {
        ui.post(() -> {
            if (closed || !waiting || token != sequence) return;
            waiting = false;
            ui.removeCallbacks(timeout);
            if (!consumed) {
                callback.setEnabled(false);
                activity.getOnBackPressedDispatcher().onBackPressed();
            }
        });
    }

    @Override public void close() {
        closed = true;
        ui.removeCallbacks(timeout);
        callback.remove();
        view.removeJavascriptInterface("smeggTunerBack");
    }
}
