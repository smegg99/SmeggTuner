//go:build android && cgo

// internal/androidhost/native_android.c
#include "native_android.h"
#include <pthread.h>
#include <stdlib.h>

static JavaVM *vm;
static jobject current_host;
static jmethodID perform;
static pthread_mutex_t host_mutex = PTHREAD_MUTEX_INITIALIZER;

void tuner_bind(JNIEnv *env, jobject host) {
    pthread_mutex_lock(&host_mutex);
    (*env)->GetJavaVM(env, &vm);
    if (current_host) (*env)->DeleteGlobalRef(env, current_host);
    current_host = (*env)->NewGlobalRef(env, host);
    jclass type = (*env)->GetObjectClass(env, host);
    perform = (*env)->GetMethodID(env, type, "perform", "(J[B)Z");
    (*env)->DeleteLocalRef(env, type);
    pthread_mutex_unlock(&host_mutex);
}

void tuner_unbind(JNIEnv *env, jobject host) {
    pthread_mutex_lock(&host_mutex);
    if (current_host && (*env)->IsSameObject(env, current_host, host)) {
        (*env)->DeleteGlobalRef(env, current_host);
        current_host = NULL;
        perform = NULL;
    }
    pthread_mutex_unlock(&host_mutex);
}

int tuner_send(jlong id, const void *data, size_t size) {
    JNIEnv *env = NULL;
    int detach = 0;
    pthread_mutex_lock(&host_mutex);
    JavaVM *active_vm = vm;
    pthread_mutex_unlock(&host_mutex);
    if (!active_vm) return 0;
    jint status = (*active_vm)->GetEnv(active_vm, (void **)&env, JNI_VERSION_1_6);
    if (status == JNI_EDETACHED) {
        if ((*active_vm)->AttachCurrentThread(active_vm, &env, NULL) != JNI_OK) return 0;
        detach = 1;
    } else if (status != JNI_OK) {
        return 0;
    }
    pthread_mutex_lock(&host_mutex);
    jobject host = current_host ? (*env)->NewLocalRef(env, current_host) : NULL;
    jmethodID method = perform;
    pthread_mutex_unlock(&host_mutex);
    int accepted = 0;
    if (host && method) {
        jbyteArray payload = (*env)->NewByteArray(env, (jsize)size);
        if (payload) {
            (*env)->SetByteArrayRegion(env, payload, 0, (jsize)size, data);
            accepted = (*env)->CallBooleanMethod(env, host, method, id, payload);
            (*env)->DeleteLocalRef(env, payload);
        }
    }
    if ((*env)->ExceptionCheck(env)) {
        (*env)->ExceptionDescribe(env);
        (*env)->ExceptionClear(env);
        accepted = 0;
    }
    if (host) (*env)->DeleteLocalRef(env, host);
    if (detach) (*active_vm)->DetachCurrentThread(active_vm);
    return accepted;
}

void *tuner_bytes(JNIEnv *env, jbyteArray data, int *size) {
    *size = data ? (*env)->GetArrayLength(env, data) : 0;
    if (!*size) return NULL;
    void *copy = malloc((size_t)*size);
    if (!copy) { *size = 0; return NULL; }
    (*env)->GetByteArrayRegion(env, data, 0, *size, copy);
    return copy;
}
