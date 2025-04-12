package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/cli/cli-plugins/metadata"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

const (
	OllamaContainerName = "mocker-model-runner"
	OllamaImage         = "ollama/ollama:latest"
	AppVersion          = "0.1.0"
)

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "model",
			Short: "Run and manage AI models",
			Long:  "Run and manage AI models using open-source tools",
		}

		// Add subcommands
		cmd.AddCommand(
			newStatusCommand(dockerCli),
			newHelpCommand(dockerCli),
			newVersionCommand(dockerCli),
			newListCommand(dockerCli),
			newPullCommand(dockerCli),
			newRmCommand(dockerCli),
			newRunCommand(dockerCli),
		)

		return cmd
	},
		metadata.Metadata{
			SchemaVersion:    "0.1.0",
			Vendor:           "Mocker",
			Version:          AppVersion,
			ShortDescription: "Run and manage AI models using open-source technologies",
			URL:              "https://github.com/richardkiene/mocker",
		})
}

// isOllamaRunning checks if the Ollama container is running
func isOllamaRunning() bool {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), OllamaContainerName)
}

// ensureOllamaRunning ensures the Ollama container is running
func ensureOllamaRunning() error {
	if !isOllamaRunning() {
		fmt.Println("Starting Mocker Model Runner...")

		// First try to remove any existing container with this name
		removeCmd := exec.Command("docker", "rm", "-f", OllamaContainerName)
		_ = removeCmd.Run() // Ignore errors if it doesn't exist

		// Create the volume if it doesn't exist
		volumeCmd := exec.Command("docker", "volume", "create", "ollama")
		_ = volumeCmd.Run() // Ignore errors if it already exists

		// Then run the container
		cmd := exec.Command(
			"docker", "run", "-d",
			"--name", OllamaContainerName,
			"-v", "ollama:/root/.ollama",
			"-p", "11434:11434",
			"--pull", "always", // Ensure image is pulled
			OllamaImage,
		)

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to start Ollama container: %w\nOutput: %s", err, string(output))
		}

		// Wait a moment for Ollama to initialize
		time.Sleep(2 * time.Second)
	}
	return nil
}

// runInOllama executes a command in the Ollama container
func runInOllama(args ...string) (string, error) {
	cmdArgs := append([]string{"exec", OllamaContainerName}, args...)
	cmd := exec.Command("docker", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// runInOllamaInteractive executes a command in the Ollama container with interactive TTY
func runInOllamaInteractive(args ...string) error {
	cmdArgs := append([]string{"exec", "-it", OllamaContainerName}, args...)
	cmd := exec.Command("docker", cmdArgs...)

	// Connect standard input, output, and error
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Status command
func newStatusCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the model runner is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			if isOllamaRunning() {
				_, _ = fmt.Fprintln(dockerCli.Out(), "Mocker Model Runner is active")
			} else {
				_, _ = fmt.Fprintln(dockerCli.Out(), "Mocker Model Runner is not running")
			}
			return nil
		},
	}
}

// Help command - displays custom help, different from the auto-generated cobra help
func newHelpCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Show the custom help",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintln(dockerCli.Out(), "Usage:  docker model COMMAND")
			_, _ = fmt.Fprintln(dockerCli.Out(), "")
			_, _ = fmt.Fprintln(dockerCli.Out(), "Commands:")
			_, _ = fmt.Fprintln(dockerCli.Out(), "  list        List models available locally")
			_, _ = fmt.Fprintln(dockerCli.Out(), "  pull        Download a model from Docker Hub")
			_, _ = fmt.Fprintln(dockerCli.Out(), "  rm          Remove a downloaded model")
			_, _ = fmt.Fprintln(dockerCli.Out(), "  run         Run a model interactively or with a prompt")
			_, _ = fmt.Fprintln(dockerCli.Out(), "  status      Check if the model runner is running")
			_, _ = fmt.Fprintln(dockerCli.Out(), "  version     Show the current version")
		},
	}
}

// Version command
func newVersionCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the current version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureOllamaRunning(); err != nil {
				return err
			}

			version, err := runInOllama("ollama", "--version")
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(dockerCli.Out(), "Mocker version: %s\nOllama version: %s", AppVersion, version)
			return nil
		},
	}
}

// getModelDetails fetches architecture and quantization details for a model
func getModelDetails(modelName string) (string, string, error) {
	output, err := runInOllama("ollama", "show", modelName)
	if err != nil {
		return "unknown", "unknown", err
	}

	// Use regex to find architecture and quantization
	archRegex := regexp.MustCompile(`architecture\s+(\S+)`)
	quantRegex := regexp.MustCompile(`quantization\s+(\S+)`)

	archMatch := archRegex.FindStringSubmatch(output)
	quantMatch := quantRegex.FindStringSubmatch(output)

	arch := "unknown"
	quant := "unknown"

	if len(archMatch) > 1 {
		arch = archMatch[1]
	}

	if len(quantMatch) > 1 {
		quant = quantMatch[1]
	}

	return arch, quant, nil
}

// List command
func newListCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List models available locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureOllamaRunning(); err != nil {
				return err
			}

			listOutput, err := runInOllama("ollama", "list")
			if err != nil {
				return err
			}

			// Process the output
			_, _ = fmt.Fprintln(dockerCli.Out(), "+MODEL       PARAMETERS  QUANTIZATION    ARCHITECTURE  MODEL ID      CREATED     SIZE")

			scanner := bufio.NewScanner(strings.NewReader(listOutput))
			// Skip header line
			if scanner.Scan() {
				_ = scanner.Text()
			}

			// Process each line
			for scanner.Scan() {
				line := scanner.Text()
				fields := strings.Fields(line)
				if len(fields) < 5 {
					continue
				}

				modelName := fields[0]
				modelID := fields[1]
				size := fields[2]
				sizeUnit := fields[3]

				// Get architecture and quantization details
				arch, quant, _ := getModelDetails(modelName)

				// Join all remaining fields for the time info
				timeIndex := 5
				if timeIndex < len(fields) {
					timeInfo := strings.Join(fields[timeIndex:], " ")

					// Estimate parameters based on size (simplified)
					var params string
					if strings.ToUpper(sizeUnit) == "GB" {
						sizeVal, _ := strconv.ParseFloat(size, 64)
						params = fmt.Sprintf("%.2f B", sizeVal*1000)
					} else {
						params = fmt.Sprintf("%.2f M", float64(parseSize(size)))
					}

					_, _ = fmt.Fprintf(dockerCli.Out(), "+%-11s %-11s %-15s %-13s %-12s %-11s %s %s\n",
						modelName, params, quant, arch, modelID, timeInfo, size, sizeUnit)
				}
			}

			return nil
		},
	}
}

// parseSize parses a size string to float
func parseSize(size string) float64 {
	val, _ := strconv.ParseFloat(size, 64)
	return val
}

// Pull command
func newPullCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "pull [model]",
		Short: "Download a model from Docker Hub",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]
			_, _ = fmt.Fprintf(dockerCli.Out(), "Pulling model %s (this is just Ollama in disguise, but don't tell anyone)...\n", modelName)

			if err := ensureOllamaRunning(); err != nil {
				return err
			}

			// Run the pull command with interactive output
			execCmd := exec.Command("docker", "exec", OllamaContainerName, "ollama", "pull", modelName)

			// Create a pipe for command output
			stdout, err := execCmd.StdoutPipe()
			if err != nil {
				return err
			}
			stderr, err := execCmd.StderrPipe()
			if err != nil {
				return err
			}

			// Start the command
			if err := execCmd.Start(); err != nil {
				return err
			}

			// Combine stdout and stderr
			outputReader := io.MultiReader(stdout, stderr)
			scanner := bufio.NewScanner(outputReader)

			// Regular expression to find size values
			sizeRegex := regexp.MustCompile(`pulling [a-f0-9]+\.\.\. 100% ▕[█▏ ]+\s+(\d+(?:\.\d+)?)\s+([KMG]B)`)

			// Collect size data
			var totalSizeKB float64

			// Process output line by line
			for scanner.Scan() {
				line := scanner.Text()
				_, _ = fmt.Fprintln(dockerCli.Out(), line)

				// Try to extract file size
				matches := sizeRegex.FindStringSubmatch(line)
				if len(matches) == 3 {
					size, _ := strconv.ParseFloat(matches[1], 64)
					unit := matches[2]

					// Convert to KB for standardization
					switch unit {
					case "MB":
						size *= 1000
					case "GB":
						size *= 1000000
					}

					totalSizeKB += size
				}
			}

			// Wait for command to finish
			if err := execCmd.Wait(); err != nil {
				return fmt.Errorf("error pulling model: %w", err)
			}

			// Display download summary
			if totalSizeKB > 1000 {
				_, _ = fmt.Fprintf(dockerCli.Out(), "Downloaded: %.2f MB\n", totalSizeKB/1000)
			} else {
				_, _ = fmt.Fprintf(dockerCli.Out(), "Downloaded: %.2f KB\n", totalSizeKB)
			}

			_, _ = fmt.Fprintf(dockerCli.Out(), "Model %s pulled successfully (just like some other tools do, but we're honest about it)\n", modelName)
			return nil
		},
	}
}

// Remove command
func newRmCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "rm [model]",
		Short: "Remove a downloaded model",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]

			if err := ensureOllamaRunning(); err != nil {
				return err
			}

			_, err := runInOllama("ollama", "rm", modelName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(dockerCli.Out(), "Model %s removed successfully (and we didn't charge you a subscription for it)\n", modelName)
			return nil
		},
	}
}

// Run command
func newRunCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "run [model] [prompt]",
		Short: "Run a model interactively or with a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]
			args = args[1:] // Remove model name from args

			if err := ensureOllamaRunning(); err != nil {
				return err
			}

			if len(args) > 0 {
				// Single prompt mode
				prompt := strings.Join(args, " ")
				_, _ = fmt.Fprintln(dockerCli.Out(), "Running with prompt (Ollama is doing all the work, but we'll take credit)...")
				// Use interactive mode to handle streaming output properly
				return runInOllamaInteractive("ollama", "run", modelName, prompt)
			} else {
				// Interactive chat mode
				_, _ = fmt.Fprintln(dockerCli.Out(), "Interactive chat mode started. Type 'Ctrl+C' to exit.")
				_, _ = fmt.Fprintln(dockerCli.Out(), "(What you're about to use is just Ollama's interface with our name on it)")
				return runInOllamaInteractive("ollama", "run", modelName)
			}
		},
	}
}
