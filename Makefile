# Define the output directory
OUTPUT_DIR=bin

# Define the platforms you want to cross-compile for
PLATFORMS= \
    linux/amd64 \
    linux/386 \
    linux/arm \
    linux/arm64 \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64 \
    windows/arm64

# Define the Go compiler
GO=go

# Define your GitHub repository
REPO_OWNER=jere-mie
REPO_NAME=meme-bot

# Extract the tag from version.txt
TAG=$(shell cat version.txt)

# Build command for each platform
define build_platform
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building for $1..."
	@GOOS=$1 GOARCH=$2 $(GO) build -o $(OUTPUT_DIR)/memebot_$1_$2$(if $(findstring windows,$1),.exe,) .
endef

.PHONY: all build release clean

all: build

build: $(PLATFORMS)

$(PLATFORMS):
	$(call build_platform,$(word 1,$(subst /, ,$@)),$(word 2,$(subst /, ,$@)))

release:
	@echo "Creating release $(TAG)..."
	@gh release create $(TAG) \
	    --repo $(REPO_OWNER)/$(REPO_NAME) \
	    --title "Release $(TAG)" \
	    --notes "Release notes for $(TAG)"
	@echo "Uploading built files to the release..."
	@for file in $(OUTPUT_DIR)/*; do \
	    gh release upload $(TAG) $$file; \
	done
	@echo "Release created successfully."

# Set execute permission for binaries
chmod:
	@chmod +x $(OUTPUT_DIR)/*

# Clean generated files
clean:
	@rm -rf $(OUTPUT_DIR)