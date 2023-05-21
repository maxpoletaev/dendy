# Dendy

Dendy is a NES/Famicom emulator written in Go and named after the soviet Famicom 
bootleg Dendy. It serves no practical purpose other than to be a toy project for
me, so it’s unlikely to beat any of the existing emulators in terms of 
performance or accuracy as well as unlikely to ever be finished.

## Status

### CPU

 * [x] Official opcodes
 * [x] Unofficial opcodes
 * [x] Runtime disassembly
 * [x] Cycle-accurate emulation
 * [ ] Accurate clock speed
 * [x] Interrupts

### Graphics

 * [x] Background rendering
 * [x] Sprite rendering
 * [ ] 8×16 sprites
 * [ ] Palettes
 * [x] Scrolling
 * [ ] Cycle-accurate emulation

### Input/Output

* [x] Graphics output
* [x] Controller 1
* [ ] Controller 2
* [ ] Zapper

### Sound

Needs research.

### Cartridges

The goal is to support top 7 mappers covering the majority of games. The
percentage indicates the number of games that use the mapper according to
nescartdb.com.

 * [x] NROM (Mapper 0) - 10%
 * [ ] MMC1 (Mapper 1) - 28%
 * [x] UxROM (Mapper 2) - 11%
 * [ ] CNROM (Mapper 3) - 6%
 * [ ] MMC3 (Mapper 4) - 24%
 * [ ] MMC5 (Mapper 5) - 1%
 * [ ] AxROM (Mapper 7) - 3%

### Test ROMs

The checked items are the ones that pass the tests completely or with minor
inaccuracies (that might be caused by the test ROMs themselves).

 * [x] Nestest CPU
 * [ ] Blargg’s CPU tests
 * [ ] Blargg’s PPU tests
 * [ ] Blargg’s APU tests

## Controller 1

```
                   ┆┆
┌───────────────────────────────────────┐
│                                       │
│    [W]                                │
│ [A]   [D]                             │
│    [S]                       [J] [K]  │
│           [Enter] [RShift]            │
│                                       │
└───────────────────────────────────────┘
```

## Resources

Although NES emulation is a pretty well-covered topic, It is still a very
interesting and challenging project to work on. Here are some of the resources
that I found particularly useful while writing this emulator. Big thanks to
everyone who made them!

### Documentation

 * [MOS 6502 CPU Reference](https://web.archive.org/web/20210429110213/http://obelisk.me.uk/6502/) by Andrew Jabobs, 2009
 * [Extra Instructions of the 65xx Series CPU](http://www.ffd2.com/fridge/docs/6502-NMOS.extra.opcodes) by Adam Vardy, 1996
 * [NES Rendering Overview](https://austinmorlan.com/posts/nes_rendering_overview/) by Austin Morlan, 2019

### Videos

 * The [NES Emulator from Scratch](nesemu) series by David Barr covers most of
   the topics from the CPU to the sound, but I found the two videos about the 
   PPU to be the most useful for understanding the obscure details of the NES
   rendering: [[1]][ppu1], [[2]][ppu2].

[nesemu]: https://www.youtube.com/playlist?list=PLrOv9FMX8xJHqMvSGB_9G9nZZ_4IgteYf
[ppu1]: https://www.youtube.com/watch?v=-THeUXqR3zY&list=PLrOv9FMX8xJHqMvSGB_9G9nZZ_4IgteYf&index=5
[ppu2]: https://www.youtube.com/watch?v=cksywUTZxlY&list=PLrOv9FMX8xJHqMvSGB_9G9nZZ_4IgteYf&index=6

### Code

During bad times, it’s always nice to look at other people’s code to see how
they solved the same problems. Here are some of the emulators written by other
people that I often referred to when I was stuck:

 * [github.com/OneLoneCoder/olcNES](https://github.com/OneLoneCoder/olcNES)
 * [github.com/ad-sho-loko/goones](https://github.com/ad-sho-loko/goones)
 * [github.com/fogleman/nes](https://github.com/fogleman/nes)
