package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"bitmerchant/tests/acceptance/screenplay"
	"bitmerchant/tests/acceptance/steps"

	"github.com/cucumber/godog"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var (
	serverCmd *exec.Cmd
	baseURL   string
	browser   *rod.Browser
)

func TestMain(m *testing.M) {
	// 1. Build the Server
	fmt.Println("Building server...")
	buildCmd := exec.Command("go", "build", "-o", "../../bin/server_test", "../../cmd/server/main.go")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to build server: %s\n%s\n", err, out)
		os.Exit(1)
	}

	// 2. Start the Server
	fmt.Println("Starting server...")
	port := "8081" // Test port
	baseURL = "http://localhost:" + port

	// Ensure binary path is absolute or correct relative to test dir
	absPath, _ := filepath.Abs("../../bin/server_test")
	serverCmd = exec.Command(absPath)
	serverCmd.Env = append(os.Environ(),
		"PORT="+port,
		"S3_BUCKET_NAME=test-bucket", // Dummy for local test
		"AWS_REGION=us-east-1",
	)
	// serverCmd.Stdout = os.Stdout // Uncomment to see server logs
	// serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		os.Exit(1)
	}

	// Wait for server to be ready
	time.Sleep(2 * time.Second) // Simple wait, ideally health check

	// 3. Launch Browser
	u := launcher.New().MustLaunch()
	browser = rod.New().ControlURL(u).MustConnect()

	// 4. Run Tests
	opts := godog.Options{
		Format:   "pretty",
		Paths:    []string{"features"},
		NoColors: true,
	}

	status := godog.TestSuite{
		Name:                "acceptance",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	// 5. Cleanup
	_ = serverCmd.Process.Kill()
	_ = browser.Close()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	featureContext := &steps.FeatureContext{
		Actors:  make(map[string]*screenplay.Actor),
		Browser: browser,
		BaseURL: baseURL,
	}

	steps.InitializeScenario(ctx, featureContext)
}

