# Touch Zoom

A small windows executable that allows you to zoom in on your screen using your mouse wheel. It emulates the touch zoom feature on modern Trackpads and Touchscreens.
Peeve of mine is that Chrome doesn't have way to map pinch-to-zoom gesture to the mouse wheel. 
There are extensions try to emulate the behavior, but they also interfere with the layout of some pages like CTRL+/- does.

This is my solution.

This started as 2 part program talking to Teensy 3.6 with touchscreen emulator running on USB.
Instead of relying on Teensy or gousb and CGO for serial communication with Teensy, I decided to use Windows API to emulate the touch screen. So I can use it on any Windows machine.

## Usage

To not interfere with my and others users daily habits, it is tucked away.
It is only active while  ```F21``` key to be pressed.

After that, it will listen to mouse wheel events (not horizontal wheel events, although it can be modified to work like that too) and emulate touch screen events.
There is a small delay between the mouse wheel event and the touch screen event, so it doesn't feel as responsive as the real thing.

I have bound ```F21``` to one of the keys on my mouse using Logi Options software. It can be bound to any key using AutoHotKey or similar software.

### Move mode

There is an experimental mode that allows you to simulate moving the screen around. Can be enabled in the systray icon menu.
I still don't like it yet, so it is disabled by default. Feel free to play with it.
(When blocking the mouse move event, windows doesn't update the mouse cursor position, so it is not possible to move the mouse cursor around the screen while in this mode. I am still looking for a solution to this problem.)

## How it works

On start, we install a global keyboard low level hook to listen to ```F21``` key press.
When the key is pressed, we install a global mouse low level hook to listen to mouse wheel events.
When the key is released, we uninstall the mouse hook.

When we receive a mouse wheel event, we emulate a touch screen event using Windows API.

## Building

This is a WIN32 api project more than a Go project. So code is messy and not idiomatic with a bunch of unsafe sprinkled around.
Also, it depends on great W32 library by Allen Dang. But since it looks abandoned, there are some compile errors in the library.
Planing to move relevant parts of the library into this project and fix the errors.

Feel free to change variables on top of the file to change the behavior of the program to your liking.

```bash
go build -ldflags -H=windowsgui -o touchzoom.exe
```
will build the executable and get rid of the console window.

## TODO

- [ ] Fix the move mode (see above)
- [ ] Add a way to change the zoom level without compiling the code
- [ ] Remove dependency on W32 library
