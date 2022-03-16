# GyroLock

Lock systemd sessions when the laptop is shake, move.

Tested on a Dell Latitude 7420.

## Install

Available as archlinux AUR package `gyrolock`. This will install also `systemd-lock-handler` AUR package and `swaylock`.

## Activate

```
sudo systemctl enable --now gyrolock.service
systemctl --user enable --now systemd-lock-handler.service
systemctl --user enable --now swaylock.service
```

## Settings

Sensitivity can be set in an systemd unit override with `SENSITIVITY` env var.
Try your own value, default is 5.

## Debug

Get sensors values with:

```
$ DEBUG=1 SENSITIVITY=5 sudo ./gyrolock
```
