# gonewire
one wire library that uses the w1 kernel module.

### current support:
```
DS18(S)20
DS1822
DS18B20
```

but should be easily extendable, don't forget to fork and PR if you do so ^^

## Usage:
Take a close look at cmd/temp-metrics ;-)

## does it run on an Raspberry Pi?
Of course! If you want to crosscompile, use:

```
env GOOS=linux GOARCH=arm GOARM=5 go build
```

Tested on an Rpi Zero W.

You might need a different GOARM value for the 4th generation.

For setting your pi, take a look at this fantastic blog post by Martin Kompf:

https://www.mkompf.com/weather/pionewiremini.html

## Bugs?

This small lib is a one nighter, but it should work under good conditions.

## I just hooked up a new sensor, but it does not show up, what should i do?

restart your program, we currently don't do autodetection here.