# TurtleSilicon 

A user-friendly launcher for Turtle WoW on Apple Silicon Macs, with one-click patching of winerosetta, rosettax87 and d9vk.

## Credits

All credit for the core translation layer `winerosetta` goes to **LifeisAwful**. This application is merely a Fyne-based GUI wrapper to simplify the patching and launching process. [https://github.com/Lifeisawful/winerosetta](https://github.com/Lifeisawful/winerosetta) 

## Features & Highlights

*   **Run 32-bit DirectX9 World of Warcraft (v1.12) on Apple Silicon:** Enjoy the classic WoW experience on your modern Mac without "illegal instruction" errors.
*   **Significant Performance Boost:**
    *   Utilizes the `rosettax87` hack by LifeisAwful to accelerate x87 FPU instructions.
    *   Integrates `d9vk` (a fork of DXVK for MoltenVK) by Kegworks-App, enabling DirectX9 to run much more efficiently on Apple Silicon via Vulkan and Metal.
    *   Experience a massive FPS increase: from around 20 FPS in unoptimized environments to **up to 200 FPS** (a 10x improvement!) in many areas.
*   **One-Click Patching:** Simplifies the setup process for both CrossOver and your Turtle WoW installation.

## Build Instructions

To build this application yourself, you will need:

1.  **Go**: Make sure you have Go installed on your system. You can download it from [golang.org](https://golang.org/).
2.  **Fyne**: Install the Fyne toolkit and its dependencies by following the instructions on the [Fyne website](https://developer.fyne.io/started/).

Once Go and Fyne are set up, navigate to the project directory in your terminal and run the following command to build the application for Apple Silicon (ARM64) macOS:

```sh
GOOS=darwin GOARCH=arm64 fyne package
```

This will create a `TurtleSilicon.app` file in the project directory, which you can then run.

Make sure you have an `Icon.png` file in the root of the project directory before building.

## Bundled Binaries

The `rosettax87` and `winerosetta` (d3d9.dll) components included in this application are precompiled for convenience. If you prefer, you can compile them yourself by following the instructions provided by LifeIsAwful on the official `winerosetta` repository: [https://github.com/Lifeisawful/winerosetta](https://github.com/Lifeisawful/winerosetta)
