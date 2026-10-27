// build/android-support/java/WebViewRequirement.java
package com.wails.app;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.ActivityNotFoundException;
import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageInfo;
import android.net.Uri;
import android.provider.Settings;
import androidx.webkit.WebViewCompat;

public final class WebViewRequirement {
    private WebViewRequirement() { }

    private static PackageInfo provider(Context context) {
        try { return WebViewCompat.getCurrentWebViewPackage(context); }
        catch (RuntimeException unavailable) { return null; }
    }

    static boolean allowed(Context context) {
        return allowed(provider(context));
    }

    private static boolean allowed(PackageInfo info) {
        if (info == null || info.versionName == null) return false;
        try {
            return Integer.parseInt(info.versionName.split("\\.")[0]) >= BuildConfig.MIN_WEBVIEW_MAJOR;
        } catch (NumberFormatException invalid) { return false; }
    }

    static boolean check(Activity activity, Runnable retry) {
        PackageInfo info = provider(activity);
        if (allowed(info)) return true;
        String description = info == null ? activity.getString(R.string.webview_unavailable)
                : info.packageName + " " + info.versionName;
        android.util.Log.i("SmeggTunerWebView", "Unsupported provider: " + description);
        AlertDialog dialog = new AlertDialog.Builder(activity)
                .setTitle(R.string.webview_title)
                .setMessage(activity.getString(R.string.webview_required, BuildConfig.MIN_WEBVIEW_MAJOR, description))
                .setPositiveButton(R.string.webview_update, null)
                .setNeutralButton(R.string.webview_retry, null)
                .setNegativeButton(R.string.webview_close, (self, which) -> activity.finish())
                .setOnCancelListener(self -> activity.finish())
                .create();
        dialog.setOnShowListener(self -> {
            dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener(view -> openUpdater(activity, info));
            dialog.getButton(AlertDialog.BUTTON_NEUTRAL).setOnClickListener(view -> {
                dialog.dismiss();
                retry.run();
            });
        });
        dialog.show();
        return false;
    }

    private static void openUpdater(Activity activity, PackageInfo info) {
        String name = info == null ? "com.google.android.webview" : info.packageName;
        try {
            activity.startActivity(new Intent(Intent.ACTION_VIEW, Uri.parse("market://details?id=" + name)));
        } catch (ActivityNotFoundException noStore) {
            activity.startActivity(new Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS, Uri.parse("package:" + name)));
        }
    }
}
