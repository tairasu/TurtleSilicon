.PHONY: build clean build-dev build-release

# Default target - optimized release build
all: build-release

# Development build (larger, with debug symbols)
build-dev:
	GOOS=darwin GOARCH=arm64 fyne package
	@echo "Copying additional resources to app bundle..."
	@mkdir -p TurtleSilicon.app/Contents/Resources/rosettax87
	@mkdir -p TurtleSilicon.app/Contents/Resources/winerosetta
	@mkdir -p TurtleSilicon.app/Contents/Resources/img/icons
	@cp -R rosettax87/* TurtleSilicon.app/Contents/Resources/rosettax87/
	@cp -R winerosetta/* TurtleSilicon.app/Contents/Resources/winerosetta/
	@cp -R img/icons/* TurtleSilicon.app/Contents/Resources/img/icons/
	@echo "Development build complete!"

build: build-release

build-release:
	@echo "Building optimized release version..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build \
		-ldflags="-s -w -X main.appVersion=$$(grep Version FyneApp.toml | cut -d'"' -f2)" \
		-trimpath \
		-tags=release \
		-o turtlesilicon .
	@echo "Packaging with fyne..."
	GOOS=darwin GOARCH=arm64 fyne package --release --executable turtlesilicon
	@echo "Copying additional resources to app bundle..."
	@mkdir -p TurtleSilicon.app/Contents/Resources/rosettax87
	@mkdir -p TurtleSilicon.app/Contents/Resources/winerosetta
	@mkdir -p TurtleSilicon.app/Contents/Resources/img/icons
	@cp -R rosettax87/* TurtleSilicon.app/Contents/Resources/rosettax87/
	@cp -R winerosetta/* TurtleSilicon.app/Contents/Resources/winerosetta/
	@cp -R img/icons/* TurtleSilicon.app/Contents/Resources/img/icons/
	@echo "Stripping additional symbols..."
	strip -x TurtleSilicon.app/Contents/MacOS/turtlesilicon
	@echo "Optimized release build complete!"
	@echo "Binary size: $$(ls -lah TurtleSilicon.app/Contents/MacOS/turtlesilicon | awk '{print $$5}')"

# Clean build artifacts
clean:
	rm -rf TurtleSilicon.app
	rm -f TurtleSilicon.dmg
	rm -f turtlesilicon

dmg: build-release
	@echo "Preparing DMG staging directory..."
	@rm -rf dmg-staging
	@mkdir dmg-staging
	@cp -R TurtleSilicon.app dmg-staging/
	@ln -s /Applications dmg-staging/Applications
	@echo "Creating DMG file..."
	@hdiutil create -volname TurtleSilicon -srcfolder dmg-staging -ov -format UDZO TurtleSilicon.dmg
	@echo "DMG created: TurtleSilicon.dmg"
	@rm -rf dmg-staging