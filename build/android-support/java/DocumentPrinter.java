// build/android-support/java/DocumentPrinter.java
package com.wails.app;

import android.app.Activity;
import android.content.Context;
import android.graphics.Bitmap;
import android.net.Uri;
import android.os.Bundle;
import android.os.CancellationSignal;
import android.os.Handler;
import android.os.Looper;
import android.os.ParcelFileDescriptor;
import android.print.PageRange;
import android.print.PrintAttributes;
import android.print.PrintDocumentAdapter;
import android.print.PrintJob;
import android.print.PrintManager;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import java.io.ByteArrayInputStream;

final class DocumentPrinter implements AutoCloseable {
    interface Completion { void finish(String value, String error); }

    private final Activity activity;
    private final Completion completion;
    private final Runnable onClose;
    private final Handler ui = new Handler(Looper.getMainLooper());
    private final Runnable timeout = () -> fail("Print document loading timed out");
    private WebView view;
    private PrintJob job;
    private boolean closed;
    private boolean started;
    private boolean replied;

    DocumentPrinter(Activity activity, Completion completion, Runnable onClose) {
        this.activity = activity;
        this.completion = completion;
        this.onClose = onClose;
    }

    @SuppressWarnings("deprecation") // API 21 uses these WebView settings and callbacks.
    void start(String html, boolean landscape, String name) {
        try {
            view = new WebView(activity);
            WebSettings settings = view.getSettings();
            settings.setJavaScriptEnabled(false);
            settings.setAllowFileAccess(false);
            settings.setAllowContentAccess(false);
            settings.setAllowFileAccessFromFileURLs(false);
            settings.setAllowUniversalAccessFromFileURLs(false);
            settings.setBlockNetworkLoads(true);
            settings.setMixedContentMode(WebSettings.MIXED_CONTENT_NEVER_ALLOW);
            settings.setCacheMode(WebSettings.LOAD_NO_CACHE);
            view.setWebViewClient(new WebViewClient() {
                @Override public boolean shouldOverrideUrlLoading(WebView web, String url) { return true; }
                @Override public WebResourceResponse shouldInterceptRequest(WebView web, WebResourceRequest request) {
                    if ("data".equals(Uri.parse(request.getUrl().toString()).getScheme())) return null;
                    return new WebResourceResponse("text/plain", "UTF-8", new ByteArrayInputStream(new byte[0]));
                }
                @Override public void onPageStarted(WebView web, String url, Bitmap icon) {
                    if (started) fail("Print document navigated during printing");
                }
                @Override public void onPageFinished(WebView web, String url) { print(landscape, name); }
                @Override public void onReceivedError(WebView web, int code, String description, String url) {
                    fail("Print document failed to load");
                }
            });
            ui.postDelayed(timeout, 60000);
            view.loadDataWithBaseURL(null, html, "text/html", "UTF-8", null);
        } catch (Exception error) {
            fail("Print initialization failed: " + error.getClass().getSimpleName());
        }
    }

    private void print(boolean landscape, String name) {
        if (closed || started) return;
        started = true;
        ui.removeCallbacks(timeout);
        try {
            PrintManager manager = (PrintManager) activity.getSystemService(Context.PRINT_SERVICE);
            if (manager == null) throw new IllegalStateException("Printing is unavailable");
            PrintDocumentAdapter delegate = view.createPrintDocumentAdapter(name);
            PrintDocumentAdapter adapter = new PrintDocumentAdapter() {
                @Override public void onStart() { delegate.onStart(); }
                @Override public void onLayout(PrintAttributes oldAttributes, PrintAttributes newAttributes,
                        CancellationSignal cancellation, LayoutResultCallback callback, Bundle extras) {
                    delegate.onLayout(oldAttributes, newAttributes, cancellation, callback, extras);
                }
                @Override public void onWrite(PageRange[] pages, ParcelFileDescriptor destination,
                        CancellationSignal cancellation, WriteResultCallback callback) {
                    delegate.onWrite(pages, destination, cancellation, callback);
                }
                @Override public void onFinish() {
                    try { delegate.onFinish(); } finally { release(); }
                }
            };
            PrintAttributes.MediaSize paper = PrintAttributes.MediaSize.ISO_A4;
            PrintAttributes attributes = new PrintAttributes.Builder()
                    .setMediaSize(landscape ? paper.asLandscape() : paper.asPortrait())
                    .setResolution(new PrintAttributes.Resolution("pdf", "PDF", 300, 300))
                    .setMinMargins(PrintAttributes.Margins.NO_MARGINS)
                    .setColorMode(PrintAttributes.COLOR_MODE_COLOR).build();
            job = manager.print(name, adapter, attributes);
            if (job == null) throw new IllegalStateException("Print job was not accepted");
            reply("", "");
        } catch (Exception error) {
            fail("Print request failed: " + error.getClass().getSimpleName());
        }
    }

    private void reply(String value, String error) {
        if (replied) return;
        replied = true;
        completion.finish(value, error);
    }

    private void fail(String error) {
        reply("", error);
        close();
    }

    private void release() {
        if (closed) return;
        closed = true;
        ui.removeCallbacks(timeout);
        if (view != null) {
            view.stopLoading();
            view.destroy();
            view = null;
        }
        onClose.run();
    }

    @Override public void close() {
        if (job != null) job.cancel();
        reply("", "Print request cancelled");
        release();
    }
}
