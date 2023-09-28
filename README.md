# Falco-Takeoff image

This is a very, very (very) crude PoC that will attempt to build the kernel module for Falco by any means necessary.

Basically, it will download the base image for your distro, spawn a container inside itself, download drivers & compilers and build the driver.

For now I PoC'd Ubuntu, Fedora and CentOS 7.

```
sudo docker run --rm -i -t --privileged -e HOST_ROOT=/host -v /etc:/host/etc:ro -v /usr:/host/usr -v /dev:/host/dev -v /proc:/host/proc lucagsd/falco-takeoff:falco-0.36.0

# takeoff build
```
