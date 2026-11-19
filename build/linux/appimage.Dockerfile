# GTK and WebKit must link against the glibc available on Ubuntu 22.04.
FROM ubuntu:22.04@sha256:0e0a0fc6d18feda9db1590da249ac93e8d5abfea8f4c3c0c849ce512b5ef8982
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        libgtk-3-dev libwebkit2gtk-4.1-dev librsvg2-dev libgirepository1.0-dev libglib2.0-bin librsvg2-common \
        dpkg-dev pkg-config binutils file findutils \
    && rm -rf /var/lib/apt/lists/*

ADD --chmod=0555 --checksum=sha256:c20cd71e3a4e3b80c3483cef793cda3f4e990aca14014d23c544ca3ce1270b4d \
    https://github.com/linuxdeploy/linuxdeploy/releases/download/1-alpha-20251107-1/linuxdeploy-x86_64.AppImage /opt/linuxdeploy.AppImage
ADD --chmod=0444 --checksum=sha256:2fca8b443c92510f1483a883f60061ad09b46b978b2631c807cd873a47ec260d \
    https://github.com/AppImage/type2-runtime/releases/download/20251108/runtime-x86_64 /opt/runtime-x86_64
