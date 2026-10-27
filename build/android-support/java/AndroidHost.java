// build/android-support/java/AndroidHost.java
package com.wails.app;

import android.app.Activity;
import android.content.ClipData;
import android.content.Intent;
import android.net.Uri;
import android.os.Handler;
import android.os.Looper;
import androidx.core.content.FileProvider;
import org.json.JSONObject;
import java.io.Closeable;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

public final class AndroidHost {
    private static final int PICK_DOCUMENT = 7340;
    private final Activity activity;
    private final Handler ui = new Handler(Looper.getMainLooper());
    private final ExecutorService copies = Executors.newSingleThreadExecutor();
    private final Map<Long, Request> pending = new HashMap<>();
    private boolean closed;
    private boolean bound;
    private Request picker;
    private Request printing;
    private DocumentPrinter printer;

    public AndroidHost(Activity activity) {
        if (activity == null) throw new IllegalArgumentException("Activity is required");
        this.activity = activity;
        System.loadLibrary("wails");
        nativeBind(this);
        bound = true;
    }

    public synchronized boolean perform(long id, byte[] request) {
        if (closed || activity.isFinishing() || activity.isDestroyed() || request == null) return false;
        byte[] payload = request.clone();
        if (id == 0) return ui.post(() -> {
            try {
                JSONObject data = new JSONObject(new String(payload, StandardCharsets.UTF_8));
                if ("cancel".equals(data.getString("kind"))) cancel(data.getLong("target"));
            } catch (Exception ignored) { /* Cancellation has no response channel. */ }
        });
        if (pending.containsKey(id)) return false;
        Request operation = new Request(id);
        pending.put(id, operation);
        if (ui.post(() -> dispatch(operation, payload))) return true;
        pending.remove(id);
        return false;
    }

    private void dispatch(Request operation, byte[] payload) {
        if (!active(operation)) return;
        if (activity.isFinishing() || activity.isDestroyed()) {
            finish(operation, "", "Activity is unavailable");
            return;
        }
        try {
            JSONObject data = new JSONObject(new String(payload, StandardCharsets.UTF_8));
            switch (data.getString("kind")) {
                case "cache": finish(operation, activity.getCacheDir().getAbsolutePath(), ""); break;
                case "save": save(operation, data); break;
                case "write": write(operation, data); break;
                case "open": open(operation, data); break;
                case "print": print(operation, data); break;
                default: finish(operation, "", "Unknown Android request kind");
            }
        } catch (Exception error) {
            finish(operation, "", "Android request failed: " + error.getClass().getSimpleName());
        }
    }

    private void save(Request operation, JSONObject data) throws Exception {
        if (picker != null) {
            finish(operation, "", "Document picker is busy");
            return;
        }
        Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT)
                .addCategory(Intent.CATEGORY_OPENABLE).setType(data.getString("mime"))
                .putExtra(Intent.EXTRA_TITLE, data.getString("name"));
        picker = operation;
        try {
            activity.startActivityForResult(intent, PICK_DOCUMENT);
        } catch (RuntimeException error) {
            picker = null;
            throw error;
        }
    }

    public boolean onActivityResult(int request, int result, Intent data) {
        if (request != PICK_DOCUMENT) return false;
        android.util.Log.d("SmeggTunerDocuments", "Document picker finished: " + result);
        Request operation = picker;
        picker = null;
        if (operation == null) return true;
        if (result != Activity.RESULT_OK) {
            finish(operation, "", "");
        } else if (data == null || data.getData() == null || !"content".equals(data.getData().getScheme())) {
            finish(operation, "", "Document picker returned no content URI");
        } else {
            finish(operation, data.getData().toString(), "");
        }
        return true;
    }

    private File privateFile(String path) throws IOException {
        File file = new File(path).getCanonicalFile();
        for (File root : new File[]{activity.getCacheDir(), activity.getFilesDir()}) {
            if (root != null && file.getPath().startsWith(root.getCanonicalPath() + File.separator)
                    && file.isFile()) return file;
        }
        throw new IOException("Input file must be under app cache or files");
    }

    private void write(Request operation, JSONObject data) throws Exception {
        String path = data.getString("path");
        Uri uri = Uri.parse(data.getString("uri"));
        if (!"content".equals(uri.getScheme())) throw new IOException("Destination must be a content URI");
        operation.future = copies.submit(() -> {
            try (InputStream input = operation.own(new FileInputStream(privateFile(path)));
                 OutputStream output = operation.own(activity.getContentResolver().openOutputStream(uri, "wt"))) {
                byte[] buffer = new byte[32768];
                int count;
                while ((count = input.read(buffer)) != -1) {
                    if (Thread.currentThread().isInterrupted() || !active(operation)) {
                        throw new IOException("Document copy cancelled");
                    }
                    output.write(buffer, 0, count);
                }
                output.flush();
            } catch (Exception error) {
                reply(operation, "", "Document write failed: " + error.getClass().getSimpleName());
                return;
            }
            reply(operation, uri.toString(), "");
        });
    }

    private void open(Request operation, JSONObject data) throws Exception {
        String value = data.getString("uri");
        Uri uri = Uri.parse(value);
        if (uri.getScheme() == null || "file".equals(uri.getScheme())) {
            String path = uri.getScheme() == null ? value : uri.getPath();
            uri = FileProvider.getUriForFile(activity, activity.getPackageName() + ".fileprovider", privateFile(path));
        } else if (!"content".equals(uri.getScheme())) {
            throw new IOException("Document must be a content URI or private file");
        }
        Intent intent = new Intent(Intent.ACTION_VIEW).setDataAndType(uri, data.getString("mime"))
                .addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        intent.setClipData(ClipData.newRawUri("", uri));
        activity.startActivity(intent);
        finish(operation, uri.toString(), "");
    }

    private void print(Request operation, JSONObject data) throws Exception {
        if (printer != null) {
            finish(operation, "", "PDF printer is busy");
            return;
        }
        String html = data.getString("html");
        boolean landscape = data.getBoolean("landscape");
        printing = operation;
        printer = new DocumentPrinter(activity, (value, error) -> finish(operation, value, error), () -> {
            printer = null;
            printing = null;
        });
        printer.start(html, landscape, data.getString("name"));
    }

    private synchronized boolean active(Request operation) {
        return !closed && pending.get(operation.id) == operation;
    }

    private void reply(Request operation, String value, String error) {
        if (!ui.post(() -> finish(operation, value, error))) finish(operation, value, error);
    }

    private synchronized void finish(Request operation, String value, String error) {
        if (pending.get(operation.id) != operation) return;
        pending.remove(operation.id);
        if (!bound) return;
        JSONObject response = new JSONObject();
        try {
            response.put("value", value);
            response.put("error", error);
        } catch (org.json.JSONException impossible) {
            throw new IllegalStateException(impossible);
        }
        nativeComplete(operation.id, response.toString().getBytes(StandardCharsets.UTF_8));
    }

    private void cancel(long target) {
        Request operation;
        synchronized (this) { operation = pending.get(target); }
        if (operation == null) return;
        operation.cancel();
        if (printing == operation && printer != null) printer.close();
        if (picker == operation) {
            try { activity.finishActivity(PICK_DOCUMENT); } catch (RuntimeException ignored) { }
        }
        finish(operation, "", "Android request cancelled");
    }

    public void close() {
        synchronized (this) {
            if (closed) return;
            closed = true;
        }
        copies.shutdownNow();
        Runnable cleanup = () -> {
            ArrayList<Request> operations;
            synchronized (this) { operations = new ArrayList<>(pending.values()); }
            for (Request operation : operations) cancel(operation.id);
            if (printer != null) printer.close();
            picker = null;
            synchronized (this) {
                if (bound) {
                    bound = false;
                    nativeUnbind(this);
                }
            }
        };
        if (Looper.myLooper() == ui.getLooper()) cleanup.run();
        else if (!ui.post(cleanup)) cleanup.run();
    }

    private static final class Request {
        final long id;
        final ArrayList<Closeable> streams = new ArrayList<>();
        volatile Future<?> future;
        private boolean cancelled;

        Request(long id) { this.id = id; }

        synchronized <T extends Closeable> T own(T stream) throws IOException {
            if (stream == null) throw new IOException("Document stream is unavailable");
            if (cancelled) {
                stream.close();
                throw new IOException("Document copy cancelled");
            }
            streams.add(stream);
            return stream;
        }

        void cancel() {
            ArrayList<Closeable> owned;
            synchronized (this) {
                cancelled = true;
                owned = new ArrayList<>(streams);
                streams.clear();
            }
            if (future != null) future.cancel(true);
            for (Closeable stream : owned) {
                try { stream.close(); } catch (Exception ignored) { }
            }
        }
    }

    private static native void nativeBind(AndroidHost host);
    private static native void nativeUnbind(AndroidHost host);
    private static native void nativeComplete(long id, byte[] response);
}
