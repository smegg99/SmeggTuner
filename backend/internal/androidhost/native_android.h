// internal/androidhost/native_android.h
#include <jni.h>
#include <stddef.h>

void tuner_bind(JNIEnv *env, jobject host);
void tuner_unbind(JNIEnv *env, jobject host);
int tuner_send(jlong id, const void *data, size_t size);
void *tuner_bytes(JNIEnv *env, jbyteArray data, int *size);
