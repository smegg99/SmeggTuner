# build/android-support/proguard-rules.pro
# Wails calls these methods by name through JNI, outside R8's Java call graph.
-keepclassmembers class com.wails.app.WailsBridge {
    public <methods>;
}

-keepclassmembers class com.wails.app.AndroidHost {
    public boolean perform(long, byte[]);
}
