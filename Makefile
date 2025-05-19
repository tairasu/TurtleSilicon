.PHONY: build clean

# Default target
all: build

# Build the application with custom resource copying
build:
	GOOS=darwin GOARCH=arm64 fyne package
	@echo "Copying additional resources to app bundle..."
	@mkdir -p TurtleSilicon.app/Contents/Resources/rosettax87
	@mkdir -p TurtleSilicon.app/Contents/Resources/winerosetta
	@cp -R rosettax87/* TurtleSilicon.app/Contents/Resources/rosettax87/
	@cp -R winerosetta/* TurtleSilicon.app/Contents/Resources/winerosetta/
	@cp -R Icon.png TurtleSilicon.app/Contents/Resources/
	@echo "Build complete!"

# Clean build artifacts
clean:
	rm -rf TurtleSilicon.app
	rm -f TurtleSilicon.dmg

# Build DMG without code signing
dmg: build
	@echo "Creating DMG file..."
	@hdiutil create -volname TurtleSilicon -srcfolder TurtleSilicon.app -ov -format UDZO TurtleSilicon.dmg
	@echo "DMG created: TurtleSilicon.dmg"
