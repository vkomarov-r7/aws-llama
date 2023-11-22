package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"text/template"

	"github.com/spf13/cobra"
)

const LAUNCHCTL_LABEL = "com.rapid7.awsllama"
const LAUNCHCTL_FILENAME = "com.rapid7.awsllama.plist"
const LAUNCHCTL_TEMPLATE = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>

    <key>Label</key>
    <string>{{ .Label }}</string>

    <key>StandardErrorPath</key>
    <string>{{ .ConfigDir }}/stderr.log</string>

    <key>StandardOutPath</key>
    <string>{{ .ConfigDir }}/stdout.log</string>

    <key>WorkingDirectory</key>
    <string>{{ .ConfigDir }}</string>

    <key>ProgramArguments</key>
    <array>
      <string>{{ .ExecutablePath }}</string>
      <string>serve</string>
    </array>

  </dict>
</plist>
`

var executablePath string

type LaunchCtlParams struct {
	Label          string
	ConfigDir      string
	ExecutablePath string
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the application as a service.",
	Long:  `Installs the application as a service. Only supports OSX. Can use launchctl to update`,
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "darwin" {
			panic("Installation is only supported on OSX.")
		}

		// Calculate the executable path.
		var err error
		if executablePath == "" {
			executablePath, err = os.Executable()
			if err != nil {
				panic(err)
			}
		}

		// Ensure all directories exist
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		launchCtlDir := path.Join(homeDir, "Library", "LaunchAgents")
		err = ensureDir(launchCtlDir)
		if err != nil {
			panic(err)
		}

		configDir := path.Join(homeDir, ".awsllama")
		err = ensureDir(configDir)
		if err != nil {
			panic(err)
		}

		// Render the template.
		params := LaunchCtlParams{
			Label:          LAUNCHCTL_LABEL,
			ConfigDir:      configDir,
			ExecutablePath: executablePath,
		}

		var tplBuffer bytes.Buffer
		tpl := template.New("launchctl.plist")
		tpl, err = tpl.Parse(LAUNCHCTL_TEMPLATE)
		if err != nil {
			panic(err)
		}
		tpl.Execute(&tplBuffer, params)

		// Write the file
		plistPath := path.Join(launchCtlDir, LAUNCHCTL_FILENAME)
		f, err := os.Create(plistPath)
		if err != nil {
			panic(err)
		}
		_, err = f.Write(tplBuffer.Bytes())
		if err != nil {
			panic(err)
		}

		// Unload any existing versions from launchctl.
		unloadCmd := exec.Command("launchctl", "unload", plistPath)
		err = unloadCmd.Run()
		if err != nil {
			fmt.Printf("Failed to unload: %s (this should be ok).\n", err)
		}

		// Load the file into launchctl.
		loadCmd := exec.Command("launchctl", "load", "-w", plistPath)
		err = loadCmd.Run()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Successfully installed application to: %s\nRun 'start' to start the system process.\n", plistPath)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	installCmd.Flags().StringVar(&executablePath, "executable", "", "Specify an alternative executable to use for the service.")
}

func ensureDir(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
