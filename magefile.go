//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName = "devrewoh-portfolio"
	dockerTag  = "devrewoh-portfolio"
)

var Default = Build

// Core Development Tasks

// Generate generates templ templates
func Generate() error {
	fmt.Println("🔄 Generating templates...")
	return sh.RunV("templ", "generate")
}

// Build builds the application binary
func Build() error {
	mg.Deps(Generate)
	fmt.Println("🔨 Building application...")
	return sh.RunV("go", "build", "-o", filepath.Join("bin", binaryName), ".")
}

// Run builds and runs the application locally
func Run() error {
	mg.Deps(Build)
	fmt.Println("🚀 Running application...")
	return sh.RunV(filepath.Join("bin", binaryName))
}

// Dev runs development server with hot reload
func Dev() error {
	fmt.Println("🚀 Starting development server...")

	// Check if air is available (use -v flag, not --version)
	if err := sh.Run("air", "-v"); err != nil {
		fmt.Println("❌ Air not found, install with: go install github.com/air-verse/air@latest")
		fmt.Println("📝 Or run 'mage install' to install all tools")
		fmt.Println("🔄 Falling back to regular run...")
		return Run()
	}

	// Kill any existing processes on port 8080
	fmt.Println("🧹 Cleaning up any existing processes...")
	if err := sh.Run("pkill", "-f", binaryName); err != nil {
		// Don't return error if no process found - that's expected
		fmt.Println("📝 No existing processes found (this is normal)")
	}

	// Kill anything specifically on port 8080
	if err := sh.Run("lsof", "-ti:8080"); err == nil {
		fmt.Println("🔌 Killing process on port 8080...")
		sh.Run("sh", "-c", "lsof -ti:8080 | xargs kill -9")
	}

	// Ensure tmp directory exists for air
	if err := os.MkdirAll("tmp", 0755); err != nil {
		fmt.Printf("⚠️  Warning: couldn't create tmp directory: %v\n", err)
	}

	// Small delay to ensure ports are released
	fmt.Println("⏳ Waiting for port cleanup...")
	sh.Run("sleep", "1")

	// Run air with the config file
	fmt.Println("🌪️  Starting Air for hot reloading...")
	return sh.RunV("air", "-c", ".air.toml")
}

// DevDebug runs development server with hot reload in debug mode
func DevDebug() error {
	fmt.Println("🚀 Starting development server in debug mode...")

	// Check if air is available
	if err := sh.Run("air", "-v"); err != nil {
		fmt.Println("❌ Air not found, install with: go install github.com/air-verse/air@latest")
		return Run()
	}

	// Kill any existing processes
	fmt.Println("🧹 Cleaning up any existing processes...")
	sh.Run("pkill", "-f", binaryName)
	sh.Run("sh", "-c", "lsof -ti:8080 | xargs kill -9")

	// Ensure tmp directory exists
	if err := os.MkdirAll("tmp", 0755); err != nil {
		fmt.Printf("⚠️  Warning: couldn't create tmp directory: %v\n", err)
	}

	// Small delay to ensure ports are released
	fmt.Println("⏳ Waiting for port cleanup...")
	sh.Run("sleep", "1")

	// Run air with debug mode
	fmt.Println("🐛 Starting Air with debug output...")
	return sh.RunV("air", "-c", ".air.toml", "-d")
}

// Watch runs templ in watch mode alongside air (alternative dev mode)
func Watch() error {
	fmt.Println("👀 Starting watch mode with templ and air...")

	// Check if air is available
	if err := sh.Run("air", "-v"); err != nil {
		fmt.Println("❌ Air not found, install with: go install github.com/air-verse/air@latest")
		return Run()
	}

	// Kill any existing processes
	fmt.Println("🧹 Cleaning up any existing processes...")
	sh.Run("pkill", "-f", binaryName)
	sh.Run("sh", "-c", "lsof -ti:8080 | xargs kill -9")

	// Ensure tmp directory exists
	if err := os.MkdirAll("tmp", 0755); err != nil {
		fmt.Printf("⚠️  Warning: couldn't create tmp directory: %v\n", err)
	}

	// Small delay to ensure ports are released
	sh.Run("sleep", "1")

	// This is an alternative approach - you might prefer this
	return sh.RunV("air")
}

// Quality Assurance

// Format formats Go code and templates
func Format() error {
	fmt.Println("🎨 Formatting code...")

	if err := sh.RunV("go", "fmt", "./..."); err != nil {
		return err
	}

	return sh.RunV("templ", "fmt", ".")
}

// Test runs tests
func Test() error {
	fmt.Println("🧪 Running tests...")
	return sh.RunV("go", "test", "-v", "./...")
}

// TestCoverage runs tests with coverage
func TestCoverage() error {
	fmt.Println("🧪 Running tests with coverage...")
	if err := sh.RunV("go", "test", "-v", "-coverprofile=coverage.out", "./..."); err != nil {
		return err
	}
	return sh.RunV("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

// Production Deployment

// BuildProd builds for production deployment
func BuildProd() error {
	mg.Deps(Generate)
	fmt.Println("🏭 Building for production...")

	env := map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
		"GOARCH":      "amd64",
	}

	return sh.RunWithV(env, "go", "build",
		"-ldflags", "-s -w",
		"-o", filepath.Join("bin", binaryName),
		".")
}

// Docker Operations

// DockerBuild builds the Docker image
func DockerBuild() error {
	fmt.Printf("🐳 Building Docker image: %s\n", dockerTag)
	return sh.RunV("docker", "build", "-t", dockerTag, ".")
}

// DockerRun runs the Docker container locally
func DockerRun() error {
	mg.Deps(DockerBuild)
	fmt.Printf("🐳 Running Docker container: %s\n", dockerTag)
	return sh.RunV("docker", "run", "-p", "8080:8080", dockerTag)
}

// Rebuild cleans and rebuilds everything
func Rebuild() error {
	fmt.Println("🔄 Rebuilding from scratch...")
	mg.Deps(Clean, Build)
	fmt.Println("✅ Rebuild completed!")
	return nil
}

// Utility Tasks

// Clean removes build artifacts
func Clean() error {
	fmt.Println("🧹 Cleaning...")

	// Remove directories and files
	paths := []string{
		"bin/",
		"tmp/",
		"coverage.out",
		"coverage.html",
	}

	for _, path := range paths {
		if err := sh.Rm(path); err != nil {
			fmt.Printf("⚠️  Warning: couldn't remove %s: %v\n", path, err)
		}
	}

	// Remove generated template files
	if err := sh.Run("find", ".", "-name", "*_templ.go", "-delete"); err != nil {
		fmt.Printf("⚠️  Warning: couldn't remove template files: %v\n", err)
	}

	fmt.Println("✅ Cleanup completed!")
	return nil
}

// Install installs required tools
func Install() error {
	fmt.Println("📦 Installing tools...")

	tools := []string{
		"github.com/a-h/templ/cmd/templ@latest",
		"github.com/air-verse/air@latest",
	}

	for _, tool := range tools {
		fmt.Printf("📦 Installing %s...\n", tool)
		if err := sh.RunV("go", "install", tool); err != nil {
			fmt.Printf("⚠️  Failed to install %s: %v\n", tool, err)
			return err
		}
		fmt.Printf("✅ Successfully installed %s\n", tool)
	}

	fmt.Println("✅ All tools installed successfully!")
	return nil
}

// Setup initializes the project
func Setup() error {
	fmt.Println("🏗️  Setting up project...")

	// Create necessary directories
	dirs := []string{"bin", "tmp", "static/css", "static/images"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("📁 Created directory: %s\n", dir)
	}

	mg.Deps(Install, Generate)

	fmt.Println("✅ Setup complete! Run 'mage dev' to start development")
	return nil
}

// Info shows project information
func Info() error {
	fmt.Println("📋 Project Info:")
	fmt.Printf("Binary: %s\n", binaryName)
	fmt.Printf("Go: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	if wd, err := os.Getwd(); err == nil {
		fmt.Printf("Directory: %s\n", wd)
	}

	// Check if tools are installed
	fmt.Println("\n🔧 Tool Status:")

	// Check templ
	if err := sh.Run("templ", "version"); err != nil {
		fmt.Printf("❌ templ: not installed\n")
	} else {
		fmt.Printf("✅ templ: installed\n")
	}

	// Check air (air uses -v not --version)
	if err := sh.Run("air", "-v"); err != nil {
		fmt.Printf("❌ air: not installed\n")
	} else {
		fmt.Printf("✅ air: installed\n")
	}

	return nil
}
