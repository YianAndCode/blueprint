.PHONY: all clean

# output directory
OUTPUT_DIR := bin
APP_NAME := blueprint

# Go build command
GO_BUILD := go build

# flags for different platforms
WINDOWS_FLAGS := GOOS=windows
LINUX_FLAGS := GOOS=linux
MACOS_FLAGS := GOOS=darwin

# architecture flags
X86_ARCH := GOARCH=386
X64_ARCH := GOARCH=amd64
ARM64_ARCH := GOARCH=arm64

# targets for each platform/architecture combination
PLATFORMS := windows_x86 windows_x64 linux_x86 linux_x64 macos_x64 macos_arm64

# Generate the target names
# TARGETS := $(addprefix blueprint_,$(PLATFORMS))

all: $(PLATFORMS)

windows_x86:
	$(WINDOWS_FLAGS) $(X86_ARCH) $(GO_BUILD) -o $(OUTPUT_DIR)/$(APP_NAME)_$@.exe .

windows_x64:
	$(WINDOWS_FLAGS) $(X64_ARCH) $(GO_BUILD) -o $(OUTPUT_DIR)/$(APP_NAME)_$@.exe .

linux_x86:
	$(LINUX_FLAGS) $(X86_ARCH) $(GO_BUILD) -o $(OUTPUT_DIR)/$(APP_NAME)_$@ .

linux_x64:
	$(LINUX_FLAGS) $(X64_ARCH) $(GO_BUILD) -o $(OUTPUT_DIR)/$(APP_NAME)_$@ .

macos_x64:
	$(MACOS_FLAGS) $(X64_ARCH) $(GO_BUILD) -o $(OUTPUT_DIR)/$(APP_NAME)_$@ .

macos_arm64:
	$(MACOS_FLAGS) $(ARM64_ARCH) $(GO_BUILD) -o $(OUTPUT_DIR)/$(APP_NAME)_$@ .

clean:
	rm -rf $(OUTPUT_DIR)
