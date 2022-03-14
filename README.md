# GyroLock

Lock systemd sessions when the laptop is shake, move.

Tested on a Dell Latitude 7420.

## Install

Available as archlinux AUR package `gyrolock`.

## Setting

Sensitivity can be set in an systemd unit override with `SENSITIVITY` env var.
Try our own value, default is 5.

## Debug

Get sensors values with:

```
$ DEBUG=1 SENSITIVITY=5 sudo ./gyrolock
```
