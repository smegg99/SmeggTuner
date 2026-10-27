package com.wails.app;

import android.Manifest;
import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.os.Build;
import android.os.Bundle;

// Check the WebView provider and request capture permission before starting Go.
public final class LauncherActivity extends Activity {
    @Override
    protected void onCreate(Bundle state) {
        super.onCreate(state);
        checkAndLaunch();
    }

    private void checkAndLaunch() {
        if (!WebViewRequirement.check(this, this::checkAndLaunch)) return;
        if (Build.VERSION.SDK_INT >= 23
                && checkSelfPermission(Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
            requestPermissions(new String[]{Manifest.permission.RECORD_AUDIO}, 1);
            return;
        }
        launch();
    }

    @Override
    public void onRequestPermissionsResult(int request, String[] permissions, int[] grants) {
        super.onRequestPermissionsResult(request, permissions, grants);
        if (request == 1) {
            // Denied capture still permits file analysis and session management.
            launch();
        }
    }

    private void launch() {
        startActivity(new Intent(this, MainActivity.class));
        finish();
    }
}
